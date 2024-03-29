## Application Name
Songs/Singers Management

## Brief Description
Usually, if a person wants to post a song or become a singer, he/she needs to sign a record company. 
With the development of blockchain technology, everyone can post your own song and become the singer without signing a record company. 
Your songs will be recorded in the blockchain network. The copyright of the song will be only own by the singer himself/herself 
instead of the record company, which is called distributing Copyrights.

## What:
Decentralized record company

## Why:
Most of the time, the copyright of the song is controlled by the record company. Also, some people lie about the sources of the songs 
and may plagiarize someone else’s ideas. With the blockchain technology, people put everything about the song into the blockchain, 
which is immutable. And since the blockchain is public, everyone in the blockchain can see the songs’ information and people can prove 
the copyright of the songs by its owner.

## How:
Singers post their songs in the blockchain. People pay for tokens to listen to music. Since music is a pure fan-driven industry, 
singers also can spend tokens letting fans advertise the songs to more people, for example, fan A may have 10 followers or 10 people 
in the peersList, fan B may have 1,000 followers or 1,000 people in the peersList, fan B should get more tokens from the singers in 
helping to advertise the songs since fan B is more influential. Or the singer can give the token holders 20 percent of the song's future earnings. 
In this way, we use the feature of token to reflect these originally soft values, since the token can be divided into very very small pieces. 
Furthermore, fans can raise money to let their idols post new songs. 

## Functionalities/Success
- Singer can post the songs’ information
- Singer receives tokens when people listen to the singer’s music
- Users spend tokens when listen to the singer’s songs
- Singer sends songs' url when confirm to receive the fan's tokens
- Users can use the url to listen to the music
- Miners can earn transaction fees
- PrivateKey/PublicKey signature authorization

## Midpoint milestone
| Milestones        | Expected Completion Date | Actual Completion Date     |
| :---              |    :----:                |          :---: |
| Design structures, includes components of the application, <br>what should be store in the songs’ info, <br>communication protocols between users, <br>algorithm of posting to blockchain| Apr 24       | Apr 25   |
| Singer posts the songs’ information   | Apr 27        | Apr 27     |
| Singer receives tokens when people listen to the music| May 3 | May 4|
| Users spend token listening to music| May 8 | May 8    |
| Miners earn transaction fees| May 10 | May 11 |
| Singer sends songs' url when confirm to receive the fan's tokens| May 12 | May 12 |
| Users can use the url to listen to the music| May 12 | May 12 |
| PrivateKey/PublicKey signature authorization | May 14 | May 14 |
| Final report and prepare project demo| May 14 | May 15 |
| Singer spends tokens letting users advertise the song| ||
| Users receive tokens when advertising the songs| ||
| Singer spends tokens when deciding to give the songs’ future profit to fans| ||
| Singer receives tokens when fans raise tokens and want their idols to make new nice songs|||

## Design
- ### Data structure
   
   - For the mpt, which stores key-value pair structure, the key would be [userID]+id+[timestamp], the value structure(jsonString on the left) would be:
   ![](images/mptStructure.png)
   
   - Here is the basic flow and process for the application
   ![](images/flow.png)
   
- ### My wallet
    How do you make it easier for users to manage their tokens and make transactions? What we need is a wallet.
    Users can store their tokens in their wallet and check the balance.
    In order to make it easier, each user will have a myWallet and here are the set up and initial values:
    - Here the initial amount of the wallet will be assumed and default setting to 100 ETH.
    - You will have your privateKey/publicKey pair in the myWallet.

- ### Node type
    A node can be a regular singer, user/listener or can be a miner.
    - singer: post songs' information, send/receive transactions, sync blocks and blockchain
    - regular user/listener: send/receive transactions, sync blocks and blockchain
    - miner: solve the puzzle, generate the new block and sync blockchain
    - both user/listener and miner: can do both things above
     
- ### Algorithm of posting it to blockchain
    - When the user make a transaction either song or payment, you will send the transaction the all the miners.
    - Miners will hold a Pending SyncTransactionPool which stores all the transactions need to be add to a new block and be confirmed.
    - Miners choose the transactions(typically will choose the one has higher transaction fee) and solve the puzzle and make the new block.
    - When the new block is added into the canonical blockchain, all the status of the transactions in that block will be confirmed and will be removed from the Pending SyncTransactionPool.
    - Miner Rewards
        - Post songs/Transaction Fees per transaction	0.5 ETH
        - Listen to the music                           1.0 ETH
        - Block Reward                                  3.0 ETH
        - Miner Reward = Block Reward + transactionsFee

## Demo
   Link: https://youtu.be/kXgE2kdHAgY
## References
[1] Inspiration from Xiaosong Gao, who is the Chairmen of Alibaba Music Group and the Chairman of Alibaba Entertainment Strategic Committee.

[2] Song Metadata Standards "https://www.synchtank.com/blog/the-guild-of-music-supervisors-suggested-metadata-standards/"

[3] Inside an Ethereum transaction "https://medium.com/@codetractio/inside-an-ethereum-transaction-fa94ffca912f"

[4] Life Cycle of an Ethereum Transaction "https://medium.com/blockchannel/life-cycle-of-an-ethereum-transaction-e5c66bae0f6e"


