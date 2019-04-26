package p1

import (
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
	"reflect"
	"strings"
)

type Flag_value struct {
	encoded_prefix []uint8 // 3: Ext, 4: Leaf
	value          string
}

type Node struct {
	node_type    int // 0: Null, 1: Branch, 2: Ext or Leaf
	branch_value [17]string
	flag_value   Flag_value
}

type MerklePatriciaTrie struct {
	db   map[string]Node
	Kv   map[string]string
	Root string
}

func (mpt *MerklePatriciaTrie) Get(key string) (string, error) {
	// 1. keyToHex
	// 2. compact_encode
	// find recursively
	hex_array := keyToHex(key)
	root := mpt.db[mpt.Root]
	value, error := mpt.findHelper(&root, hex_array)

	return value, error
}

func (mpt *MerklePatriciaTrie) findHelper(node *Node, input_array []uint8) (string, error) {
	n_type := getType(node) // 0 null, 1 branch, 3 extension 4 leaf
	if n_type != 1 {
		node_prefix := compact_decode(node.flag_value.encoded_prefix)
		common_path, input_rest, current_rest := splitPath(node_prefix, input_array)

		// if node is a leaf
		if n_type == 4 {
			// len(key) == 0 len (leaf.prefix) == 0 -> ext -> leaf / branch -> leaf
			if isSliceEqual(input_rest, current_rest) {
				return node.flag_value.value, nil
			}
			return "", errors.New("path_not_found")
		}

		// if node is a extension node
		if n_type == 3 {
			if len(common_path) == 0 {
				return "", errors.New("path_not_found")
			}
			temp := mpt.db[node.flag_value.value]
			return mpt.findHelper(&temp, input_rest)
		}
	}

	// branch node
	if n_type == 1 {
		// ext -> branch / not exists
		if len(input_array) == 0 {
			if node.branch_value[16] != "" {
				return node.branch_value[16], nil
			}
			return "", errors.New("path_not_found")
		}
		// no next pointer -> not exists
		if node.branch_value[input_array[0]] == "" {
			return "", errors.New("path_not_found")
		}
		temp := mpt.db[node.branch_value[input_array[0]]]
		return mpt.findHelper(&temp, input_array[1:])
	}
	return "", errors.New("path_not_found")
}

func (mpt *MerklePatriciaTrie) Insert(key string, new_value string) {
	hex_array := keyToHex(key)

	if mpt.Root == "" {
		var new_node Node
		hex_array = append(hex_array, 16)
		encoded_prefix := compact_encode(hex_array)

		flag_value := Flag_value{encoded_prefix, new_value}
		new_node.node_type = 2
		new_node.flag_value = flag_value
		hash_value := new_node.hash_node()

		mpt.db[hash_value] = new_node
		mpt.Root = hash_value
	} else {
		temp := mpt.db[mpt.Root]
		mpt.Root = mpt.insertHelper(&temp, hex_array, new_value)
	}

	mpt.Kv[key] = new_value
}

func (mpt *MerklePatriciaTrie) insertHelper(node *Node, path_array []uint8, new_value string) string {
	n_type := getType(node) // 0 null, 1 branch, 3 extension 4 leaf

	if n_type == 1 {
		input_rest := make([]uint8, 0)

		if len(path_array) == 0 {
			node.branch_value[16] = new_value

			hash := mpt.updateMap(node)
			return hash
		}
		branch_index := path_array[0]
		input_rest = append(input_rest, path_array...)

		if node.branch_value[branch_index] == "" { //  branch[i] == "", branch[i] and leaf
			input_rest = append(input_rest, 16)
			new_leaf_hash_value := mpt.addNewLeaf(input_rest[1:], new_value)
			node.branch_value[branch_index] = new_leaf_hash_value
			hash := mpt.updateMap(node)

			return hash
		} else { //  branch[i] != "", branch[i] and insertHelper
			temp := mpt.db[node.branch_value[branch_index]]
			new_node_hash_value := mpt.insertHelper(&temp, input_rest[1:], new_value)

			// delete(mpt.db, node.branch_value[branch_index])
			node.branch_value[branch_index] = new_node_hash_value
			hash := mpt.updateMap(node)
			return hash
		}
	}

	if n_type == 3 {
		// decoded, without 16 without prefix
		node_prefix := compact_decode(node.flag_value.encoded_prefix)
		common_path, input_rest, current_rest := splitPath(node_prefix, path_array)

		if len(path_array) == 0 {
			// change current(extension) -> new(branch)
			var branch Node
			branch.node_type = 1
			branch.branch_value[16] = new_value

			if len(current_rest) > 1 {
				ext_hash := mpt.addNewExt(current_rest[1:], node.flag_value.value)
				branch.branch_value[current_rest[0]] = ext_hash
			} else {
				branch.branch_value[current_rest[0]] = node.flag_value.value
			}
			hash := mpt.updateMap(&branch)
			return hash
		}

		if len(current_rest) == 0 {
			temp := mpt.db[node.flag_value.value]
			hash := mpt.insertHelper(&temp, input_rest, new_value)

			node.flag_value.value = hash
			ext_hash := mpt.updateMap(node)
			return ext_hash
		}

		if len(common_path) == 0 {
			var branch Node
			branch.node_type = 1

			if len(current_rest) > 1 {
				ext_hash := mpt.addNewExt(current_rest[1:], node.flag_value.value)
				branch.branch_value[current_rest[0]] = ext_hash
			} else {
				branch.branch_value[current_rest[0]] = node.flag_value.value
			}
			input_rest = append(input_rest, 16)
			leaf_hash := mpt.addNewLeaf(input_rest[1:], new_value)
			branch.branch_value[input_rest[0]] = leaf_hash
			hash := mpt.updateMap(&branch)
			return hash
		}

		if len(common_path) != 0 {
			if len(current_rest) == 0 && len(input_rest) == 0 {
				temp := mpt.db[node.flag_value.value]
				hash := mpt.insertHelper(&temp, input_rest, new_value)
				node.flag_value.value = hash
				ext_hash := mpt.updateMap(node)
				return ext_hash
			}
			if len(current_rest) != 0 && len(input_rest) != 0 {
				encoded_path := compact_encode(common_path)
				node.flag_value.encoded_prefix = encoded_path

				var branch Node
				branch.node_type = 1
				if len(current_rest) > 1 {
					ext_hash := mpt.addNewExt(current_rest[1:], node.flag_value.value)
					branch.branch_value[current_rest[0]] = ext_hash
				} else {
					branch.branch_value[current_rest[0]] = node.flag_value.value
				}
				input_rest = append(input_rest, 16)
				leaf_hash := mpt.addNewLeaf(input_rest[1:], new_value)
				branch.branch_value[input_rest[0]] = leaf_hash

				hash := mpt.updateMap(&branch)
				node.flag_value.value = hash
				ext_hash := mpt.updateMap(node)
				return ext_hash
			}

			if len(current_rest) == 0 {
				temp := mpt.db[node.flag_value.value]
				hash := mpt.insertHelper(&temp, input_rest, new_value)

				node.flag_value.value = hash
				ext_hash := mpt.updateMap(node)
				return ext_hash
			}

			if len(input_rest) == 0 {
				node.flag_value.encoded_prefix = compact_encode(common_path)
				var branch Node
				branch.node_type = 1
				branch.branch_value[current_rest[0]] = node.flag_value.value
				branch.branch_value[16] = new_value

				ext_hash := mpt.addNewExt(current_rest[1:], node.flag_value.value)
				branch.branch_value[current_rest[0]] = ext_hash
				branch_hash := mpt.updateMap(&branch)

				node.flag_value.value = branch_hash
				cur_hash := mpt.updateMap(node)
				return cur_hash
			}
		}
	}
	// leaf node
	if n_type == 4 {
		// decoded, without 16 without prefix
		node_prefix := compact_decode(node.flag_value.encoded_prefix)
		common_path, input_rest, current_rest := splitPath(node_prefix, path_array)

		// if it is the same hash value
		if isSliceEqual(node_prefix, path_array) {
			node.flag_value.value = new_value
			hash := mpt.updateMap(node)
			return hash
		}

		if len(common_path) == 0 {
			if len(input_rest) == 0 && len(current_rest) != 0 {
				var branch Node
				branch.node_type = 1
				branch.branch_value[16] = new_value

				path := make([]uint8, 0)
				path = append(path, current_rest[1:]...)
				path = append(path, 16)

				leaf_hash := mpt.addNewLeaf(path, node.flag_value.value)
				branch.branch_value[current_rest[0]] = leaf_hash
				branch_hash := mpt.updateMap(&branch)

				return branch_hash
			}
			if len(current_rest) == 0 {
				input_rest = append(input_rest, 16)

				var new_branch Node
				new_branch.node_type = 1
				branch_index := input_rest[0]
				new_leaf_hash_value := mpt.addNewLeaf(input_rest[1:], new_value)
				new_branch.branch_value[branch_index] = new_leaf_hash_value
				new_branch.branch_value[16] = node.flag_value.value
				hash := mpt.updateMap(&new_branch)

				return hash
			}

			if len(node_prefix) == 1 && node_prefix[0] == uint8(0) {
				input_rest = append(input_rest, 16)

				var new_branch Node
				new_branch.node_type = 1
				branch_index := input_rest[0]

				new_leaf_hash_value := mpt.addNewLeaf(input_rest[1:], new_value)
				new_branch.branch_value[branch_index] = new_leaf_hash_value
				new_branch.branch_value[16] = node.flag_value.value

				hash := mpt.updateMap(&new_branch)
				return hash
			}

			if len(path_array) == 0 {
				input_rest = append(input_rest, 16)
				var new_branch Node
				new_branch.node_type = 1
				branch_index := input_rest[0]

				new_leaf_hash_value := mpt.addNewLeaf(input_rest[1:], node.flag_value.value)
				new_branch.branch_value[branch_index] = new_leaf_hash_value
				new_branch.branch_value[16] = new_value

				hash := mpt.updateMap(&new_branch)
				return hash
			}

			input_rest = append(input_rest, 16)
			current_rest = append(current_rest, 16)

			var new_branch Node
			new_branch.node_type = 1
			branch_index1 := input_rest[0]
			branch_index2 := current_rest[0]

			new_leaf1_hash_value := mpt.addNewLeaf(input_rest[1:], new_value)
			new_leaf2_hash_value := mpt.addNewLeaf(current_rest[1:], node.flag_value.value)

			new_branch.branch_value[branch_index1] = new_leaf1_hash_value
			new_branch.branch_value[branch_index2] = new_leaf2_hash_value

			hash := mpt.updateMap(&new_branch)
			return hash
		}

		if len(current_rest) == 0 {
			input_rest = append(input_rest, 16)

			var new_branch Node
			new_branch.node_type = 1
			branch_index := input_rest[0]

			new_leaf_hash_value := mpt.addNewLeaf(input_rest[1:], new_value)

			new_branch.branch_value[branch_index] = new_leaf_hash_value
			new_branch.branch_value[16] = node.flag_value.value
			hash := mpt.updateMap(&new_branch)

			new_ext_hash_value := mpt.addNewExt(common_path, hash)
			return new_ext_hash_value
		}

		if len(common_path) == len(path_array) {
			current_rest = append(current_rest, 16)

			var new_branch Node
			new_branch.node_type = 1
			branch_index := current_rest[0]

			new_leaf_hash_value := mpt.addNewLeaf(current_rest[1:], node.flag_value.value)
			new_branch.branch_value[branch_index] = new_leaf_hash_value
			new_branch.branch_value[16] = new_value

			hash := mpt.updateMap(&new_branch)
			new_ext_hash_value := mpt.addNewExt(common_path, hash)
			return new_ext_hash_value
		}

		if len(common_path) < len(node_prefix) && len(common_path) < len(path_array) {
			current_rest = append(current_rest, 16)
			input_rest = append(input_rest, 16)

			var new_branch Node
			new_branch.node_type = 1
			branch_index1 := current_rest[0]
			branch_index2 := input_rest[0]

			new_leaf1_hash_value := mpt.addNewLeaf(current_rest[1:], node.flag_value.value)
			new_leaf2_hash_value := mpt.addNewLeaf(input_rest[1:], new_value)
			new_branch.branch_value[branch_index1] = new_leaf1_hash_value
			new_branch.branch_value[branch_index2] = new_leaf2_hash_value
			hash := mpt.updateMap(&new_branch)
			new_ext_hash_value := mpt.addNewExt(common_path, hash)
			return new_ext_hash_value
		}
	}
	return ""
}

func (mpt *MerklePatriciaTrie) updateMap(node *Node) string {
	hash := node.hash_node()
	mpt.db[hash] = *node
	return hash
}

func splitPath(current []uint8, input []uint8) ([]uint8, []uint8, []uint8) {
	common_path := make([]uint8, 0)
	current_rest := make([]uint8, 0)
	input_rest := make([]uint8, 0)

	index := 0
	for index < len(current) && index < len(input) && current[index] == input[index] {
		common_path = append(common_path, current[index])
		index++
	}

	i := 0
	current_index := index
	input_index := index
	for current_index < len(current) {
		current_rest = append(current_rest, current[current_index])
		current_index++
		i++
	}

	for input_index < len(input) {
		input_rest = append(input_rest, input[input_index])
		input_index++
		i++
	}

	return common_path, input_rest, current_rest
}

func isSliceEqual(array1 []uint8, array2 []uint8) bool {
	if len(array1) != len(array2) {
		return false
	}

	m_len := len(array1)
	for i := 0; i < m_len; i++ {
		if array1[i] != array2[i] {
			return false
		}
	}
	return true
}

func getType(node *Node) int {
	if node.node_type < 2 {
		return node.node_type
	}

	prefix := node.flag_value.encoded_prefix
	if prefix[0]/16 < 2 {
		return 3
	}
	return 4
}

func (mpt *MerklePatriciaTrie) addNewLeaf(prefix_array []uint8, value string) string {
	var new_leaf Node
	new_leaf.node_type = 2
	new_leaf_encoded_prefix := compact_encode(prefix_array)
	new_leaf.flag_value = Flag_value{new_leaf_encoded_prefix, value}

	hash := mpt.updateMap(&new_leaf)
	return hash
}

func (mpt *MerklePatriciaTrie) addNewExt(prefix_array []uint8, value string) string {
	var new_ext Node
	new_ext.node_type = 2
	new_ext_encoded_prefix := compact_encode(prefix_array)
	new_ext.flag_value = Flag_value{new_ext_encoded_prefix, value}

	hash := mpt.updateMap(&new_ext)
	return hash
}

func keyToHex(key string) []uint8 {
	byte_array := []byte(key)
	nibbles := asciiToHex(byte_array)

	return nibbles
}

func (mpt *MerklePatriciaTrie) Delete(key string) (string, error) {
	result, error := mpt.Get(key)
	_ = result

	if error == nil {
		root := mpt.db[mpt.Root]
		hex_array := keyToHex(key)
		hash, prefix, value, error := mpt.deleteHelper(&root, hex_array)

		if hash == "" {
			prefix = append(prefix, 16)
			leaf_hash := mpt.addNewLeaf(prefix, value)

			mpt.Root = leaf_hash
		} else {
			mpt.Root = hash
		}
		_ = prefix

		root = mpt.db[mpt.Root]
		n_type := getType(&root)
		if n_type == 3 {
			value, ok := mpt.db[root.flag_value.value]
			_ = value
			if ok == false {
				path := compact_decode(root.flag_value.encoded_prefix)
				path = append(path, 16)
				leaf_hash := mpt.addNewLeaf(path, root.flag_value.value)
				mpt.Root = leaf_hash
			}
		}

		delete(mpt.Kv, key)
		return value, error
	}
	return "", errors.New("path_not_found")
}

func (mpt *MerklePatriciaTrie) deleteHelper(node *Node, key []uint8) (string, []uint8, string, error) {
	n_type := getType(node) // 0 null, 1 branch, 3 extension 4 leaf

	if n_type != 1 {
		node_prefix := compact_decode(node.flag_value.encoded_prefix)
		common_path, input_rest, current_rest := splitPath(node_prefix, key)
		// if node is a leaf
		if n_type == 4 {
			// len(key) == 0 len (leaf.prefix) == 0 -> branch -> leaf
			if isSliceEqual(input_rest, current_rest) {
				return "", nil, "", nil
			}
			return "", nil, "", errors.New("path_not_found")
		}

		if n_type == 3 {
			if len(common_path) == 0 {
				return "", nil, "", errors.New("path_not_found")
			}

			temp := mpt.db[node.flag_value.value]
			hash, prefix, value, error := mpt.deleteHelper(&temp, input_rest)
			_ = error

			if len(prefix) != 0 {
				if hash != "" {
					node_prefix = append(node_prefix, prefix...)
					node_prefix = append(node_prefix, 16)
					leaf_hash := mpt.addNewLeaf(node_prefix, value)
					return leaf_hash, nil, "", nil
				}
				node_prefix = append(node_prefix, prefix...)
				ext_hash := mpt.addNewExt(node_prefix, value)
				return ext_hash, nil, "", nil
			}

			if hash != "" {
				node.flag_value.value = hash
				ext_hash := mpt.updateMap(node)
				return ext_hash, nil, "", nil
			}

			if value != "" {
				node_prefix = append(node_prefix, 16)
				leaf_hash := mpt.addNewLeaf(node_prefix, value)
				return leaf_hash, nil, "", nil
			}
		}
	}
	if n_type == 1 {
		// ext -> branch / not exists
		if len(key) == 0 {
			if node.branch_value[16] != "" {
				node.branch_value[16] = ""

				if countNotEmpty(node) == 1 {
					var index uint8
					var last Node
					for i := 0; i < 16; i++ {
						if node.branch_value[i] != "" {
							last = mpt.db[node.branch_value[i]]
							index = uint8(i)
						}
					}
					// compress
					var prefix = make([]uint8, 0)
					var compress_value string
					if getType(&last) == 3 || getType(&last) == 4 {
						prefix = compact_decode(last.flag_value.encoded_prefix)
						compress_value = last.flag_value.value
						current_path := make([]uint8, 0)
						current_path = append(current_path, index)
						current_path = append(current_path, prefix...)
						return "", current_path, compress_value, nil
					}
					new_path := make([]uint8, 0)
					new_path = append(new_path, index)

					return "", new_path, node.branch_value[index], nil
				}
				branch_hash := mpt.updateMap(node)
				return branch_hash, nil, "", nil
			}
			return "", nil, "", errors.New("path_not_found")
		}
		// no next pointer -> not exists
		if node.branch_value[key[0]] == "" {
			return "", nil, "", errors.New("path_not_found")
		}
		temp := mpt.db[node.branch_value[key[0]]]
		hash, prefix, compress_value, error := mpt.deleteHelper(&temp, key[1:])
		_ = error

		if hash == "" && len(prefix) == 0 && compress_value == "" {
			node.branch_value[key[0]] = ""
		}

		// two nodes, hash == "" -> one node
		// 		find last node
		// 		compression
		// two nodes, hash != "" -> two nodes
		// 		update
		if countNotEmpty(node) > 1 {
			if len(prefix) != 0 {
				value, ok := mpt.db[compress_value]
				_ = value

				if ok == false {
					prefix = append(prefix, 16)
					leaf_hash := mpt.addNewLeaf(prefix, compress_value)
					node.branch_value[key[0]] = leaf_hash
					branch_hash := mpt.updateMap(node)
					return branch_hash, nil, "", nil
				}

				ext_hash := mpt.addNewExt(prefix, compress_value)
				node.branch_value[key[0]] = ext_hash
				branch_hash := mpt.updateMap(node)
				return branch_hash, nil, "", nil
			}
			if len(prefix) == 0 && compress_value != "" {
				prefix = append(prefix, 16)
				leaf_hash := mpt.addNewLeaf(prefix, compress_value)
				node.branch_value[key[0]] = leaf_hash
				branch_hash := mpt.updateMap(node)
				return branch_hash, nil, "", nil
			}
			node.branch_value[key[0]] = hash
			branch_hash := mpt.updateMap(node)
			return branch_hash, nil, "", nil
		}

		if countNotEmpty(node) == 1 {
			if hash == "" {
				var index uint8
				for i := 0; i < 16; i++ {
					if node.branch_value[i] != "" {
						temp = mpt.db[node.branch_value[i]]
						index = uint8(i)
					}
				}

				if node.branch_value[16] == "" {
					// compress
					var prefix = make([]uint8, 0)
					var compress_value string
					if getType(&temp) == 3 || getType(&temp) == 4 {
						prefix = compact_decode(temp.flag_value.encoded_prefix)
						compress_value = temp.flag_value.value
						current_path := make([]uint8, 0)
						current_path = append(current_path, index)
						current_path = append(current_path, prefix...)
						return "", current_path, compress_value, nil
					}
					new_path := make([]uint8, 0)
					new_path = append(new_path, index)

					return "", new_path, node.branch_value[index], nil
				}
			}
			node.branch_value[key[0]] = hash
			branch_hash := mpt.updateMap(node)
			return branch_hash, nil, "", nil
		}

		if countNotEmpty(node) == 0 {
			if hash == "" {
				if node.branch_value[16] != "" {
					return "", nil, node.branch_value[16], nil
				}
			}

			if node.branch_value[16] != "" {
				node.branch_value[key[0]] = hash

				prefix = append(prefix, 16)
				leaf_hash := mpt.addNewLeaf(prefix, compress_value)
				node.branch_value[key[0]] = leaf_hash
				branch_hash := mpt.updateMap(node)

				return branch_hash, nil, "", nil
			}
			// compress
			current_path := make([]uint8, 0)
			current_path = append(current_path, key[0])
			current_path = append(current_path, prefix...)
			return "", current_path, compress_value, nil
		}

		branch_hash := mpt.updateMap(node)
		return branch_hash, nil, "", nil
	}

	return "", nil, "", errors.New("path_not_found")
}

func countNotEmpty(node *Node) int {
	n_type := getType(node)
	count := 0
	if n_type == 1 {
		for i := 0; i < 16; i++ {
			if node.branch_value[i] != "" {
				count++
			}
		}
		return count
	}
	return -1
}

func compact_encode(hex_array []uint8) []uint8 {
	term := 0
	if len(hex_array) > 0 && hex_array[len(hex_array)-1] == 16 {
		term = 1
		hex_array = append(hex_array[:len(hex_array)-1])
	}

	oddlen := uint8(len(hex_array) % 2)
	flags := uint8(2*term) + oddlen

	flag_hex_array := make([]uint8, 0)
	if oddlen == 1 {
		flag_hex_array = append(flag_hex_array, flags)
	} else {
		flag_hex_array = append(flag_hex_array, flags, 0)
	}
	flag_hex_array = append(flag_hex_array, hex_array...)

	ascii_array := make([]uint8, len(flag_hex_array)/2)
	for i, j := 0, 0; i < len(flag_hex_array); i, j = i+2, j+1 {
		ascii_array[j] = flag_hex_array[i]*16 + flag_hex_array[i+1]
	}

	return ascii_array
}

// If Leaf, ignore 16 at the end
func compact_decode(encoded_arr []uint8) []uint8 {
	base := asciiToHex(encoded_arr)
	begin := 2 - base[0]&1

	return base[begin:]
}

func asciiToHex(arr []uint8) []uint8 {
	l := len(arr) * 2
	var nibbles = make([]uint8, l)
	for i, b := range arr {
		nibbles[i*2] = b / 16
		nibbles[i*2+1] = b % 16
	}

	return nibbles
}

func test_compact_encode() {
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{1, 2, 3, 4, 5})), []uint8{1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 1, 2, 3, 4, 5})), []uint8{0, 1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 15, 1, 12, 11, 8, 16})), []uint8{0, 15, 1, 12, 11, 8}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{15, 1, 12, 11, 8, 16})), []uint8{15, 1, 12, 11, 8}))
}

func (node *Node) hash_node() string {
	var str string
	switch node.node_type {
	case 0:
		str = ""
	case 1:
		str = "branch_"
		for _, v := range node.branch_value {
			str += v
		}
	case 2:
		str = node.flag_value.value
	}

	sum := sha3.Sum256([]byte(str))
	return "HashStart_" + hex.EncodeToString(sum[:]) + "_HashEnd"
}

func (node *Node) String() string {
	str := "empty string"
	switch node.node_type {
	case 0:
		str = "[Null Node]"
	case 1:
		str = "Branch["
		for i, v := range node.branch_value[:16] {
			str += fmt.Sprintf("%d=\"%s\", ", i, v)
		}
		str += fmt.Sprintf("value=%s]", node.branch_value[16])
	case 2:
		encoded_prefix := node.flag_value.encoded_prefix
		node_name := "Leaf"
		if is_ext_node(encoded_prefix) {
			node_name = "Ext"
		}
		ori_prefix := strings.Replace(fmt.Sprint(compact_decode(encoded_prefix)), " ", ", ", -1)
		str = fmt.Sprintf("%s<%v, value=\"%s\">", node_name, ori_prefix, node.flag_value.value)
	}
	return str
}

func node_to_string(node Node) string {
	return node.String()
}

func (mpt *MerklePatriciaTrie) Initial() {
	mpt.db = make(map[string]Node)
	mpt.Kv = make(map[string]string)
	mpt.Root = ""
}

func is_ext_node(encoded_arr []uint8) bool {
	return encoded_arr[0]/16 < 2
}

func TestCompact() {
	test_compact_encode()
}

func (mpt *MerklePatriciaTrie) String() string {
	content := fmt.Sprintf("ROOT=%s\n", mpt.Root)
	for hash := range mpt.db {
		content += fmt.Sprintf("%s: %s\n", hash, node_to_string(mpt.db[hash]))
	}
	return content
}

func (mpt *MerklePatriciaTrie) Order_nodes() string {
	raw_content := mpt.String()
	content := strings.Split(raw_content, "\n")
	root_hash := strings.Split(strings.Split(content[0], "HashStart")[1], "HashEnd")[0]
	queue := []string{root_hash}
	i := -1
	rs := ""
	cur_hash := ""
	for len(queue) != 0 {
		last_index := len(queue) - 1
		cur_hash, queue = queue[last_index], queue[:last_index]
		i += 1
		line := ""
		for _, each := range content {
			if strings.HasPrefix(each, "HashStart"+cur_hash+"HashEnd") {
				line = strings.Split(each, "HashEnd: ")[1]
				rs += each + "\n"
				rs = strings.Replace(rs, "HashStart"+cur_hash+"HashEnd", fmt.Sprintf("Hash%v", i), -1)
			}
		}
		temp2 := strings.Split(line, "HashStart")
		flag := true
		for _, each := range temp2 {
			if flag {
				flag = false
				continue
			}
			queue = append(queue, strings.Split(each, "HashEnd")[0])
		}
	}
	return rs
}

func Test() {
	fmt.Println(">>>>>>>>>>>>>>>>>>>")
	var mpt MerklePatriciaTrie

	mpt.Initial()
	mpt.Insert("aab", "apple")
	mpt.Insert("app", "banana")
	mpt.Insert("ace", "new")
	fmt.Println(mpt.Order_nodes())
	mpt.Delete("c")
	fmt.Println(mpt.Order_nodes())
	mpt.Delete("ace")
	fmt.Println(mpt.Order_nodes())
}
