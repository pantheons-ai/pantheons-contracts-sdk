package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pantheons-ai/sdk-go/config"
	"github.com/pantheons-ai/sdk-go/pkg/pantheon"
	"log"
	"math/big"
	"strings"
)

func main() {
	// 加载配置文件
	cfg, err := config.LoadConfig("config/config.yml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化区块链客户端
	client, err := ethclient.Dial(cfg.RPCURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	// 加载私钥
	privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	// 获取链ID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get network ID: %v", err)
	}

	// 创建交易签名者
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatalf("Failed to create transaction signer: %v", err)
	}

	// 设置合约地址并创建合约实例
	contractAddress := common.HexToAddress(cfg.ContractAddress)
	instance, err := pantheon.NewPantheon(contractAddress, client)
	if err != nil {
		log.Fatalf("Failed to instantiate a Pantheon contract: %v", err)
	}

	// 生成一个新的随机私钥（新用户）
	newPrivateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("Failed to generate random private key: %v", err)
	}

	// 从私钥中获取公钥
	newPublicKey := newPrivateKey.Public()
	newPublicKeyECDSA, ok := newPublicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Error casting public key to ECDSA")
	}

	// 从公钥获取新地址
	newAddress := crypto.PubkeyToAddress(*newPublicKeyECDSA)
	fmt.Printf("New random address generated: %s\n", newAddress.Hex())

	// 调用addToWhitelist方法
	tx, err := instance.AddToWhitelist(auth, newAddress)
	if err != nil {
		log.Fatalf("Failed to invoke addToWhitelist: %v", err)
	}

	// 输出交易哈希
	fmt.Printf("Whitelist add transaction sent: %s\n", tx.Hash().Hex())

	// 等待交易被挖掘
	fmt.Println("Waiting for transaction to be mined...")
	bind.WaitMined(context.Background(), client, tx)

	// 验证地址是否已经加入白名单
	isWhitelisted, err := instance.IsWhitelisted(&bind.CallOpts{}, newAddress)
	if err != nil {
		log.Fatalf("Failed to invoke isWhitelisted: %v", err)
	}

	// 输出验证结果
	fmt.Printf("Address %s whitelisted status: %v\n", newAddress.Hex(), isWhitelisted)

	// 创建ERC404实例
	tx, err = instance.CreateERC404(auth, "TestToken", "TTK", 18, big.NewInt(0), auth.From)
	if err != nil {
		log.Fatalf("Failed to create ERC404: %v", err)
	}
	fmt.Printf("Create ERC404 transaction sent: %s\n", tx.Hash().Hex())

	// 等待交易被挖掘
	fmt.Println("Waiting for transaction to be mined...")
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatalf("Failed to mine CreateERC404 transaction: %v", err)
	}

	// 过滤ERC404Created事件
	blockNumber := receipt.BlockNumber.Uint64()
	eventIterator, err := instance.FilterERC404Created(&bind.FilterOpts{Start: blockNumber, End: &blockNumber}, nil, nil)
	if err != nil {
		log.Fatalf("Failed to filter ERC404Created events: %v", err)
	}
	defer eventIterator.Close()

	// 获取事件
	if eventIterator.Next() {
		event := eventIterator.Event
		fmt.Printf("ERC404Created event received: id=%s, contractAddress=%s\n", event.Id.String(), event.ContractAddress.Hex())
	} else if eventIterator.Error() != nil {
		log.Fatalf("Error during iteration: %v", err)
	} else {
		log.Fatal("ERC404Created event not found in the transaction receipt")
	}

	// 添加CIDs
	cids := []string{"cid1", "cid2", "cid3"}
	tx, err = instance.AddCIDs(auth, big.NewInt(0), auth.From, cids)
	if err != nil {
		log.Fatalf("Failed to add CIDs: %v", err)
	}
	fmt.Printf("Add CIDs transaction sent: %s\n", tx.Hash().Hex())

	// 等待交易被挖掘
	fmt.Println("Waiting for transaction to be mined...")
	bind.WaitMined(context.Background(), client, tx)

	// 查询贡献
	contribution, err := instance.GetContribution(&bind.CallOpts{}, big.NewInt(0), auth.From)
	if err != nil {
		log.Fatalf("Failed to get contribution: %v", err)
	}
	fmt.Printf("Contribution: %s\n", contribution.String())

	// 查询CID列表
	storedCIDs, err := instance.GetCIDs(&bind.CallOpts{}, big.NewInt(0), auth.From)
	if err != nil {
		log.Fatalf("Failed to get CIDs: %v", err)
	}
	fmt.Printf("CIDs: %s\n", strings.Join(storedCIDs, ", "))

}
