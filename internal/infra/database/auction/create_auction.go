package auction

import (
	"context"
	"fmt"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}
type AuctionRepository struct {
	Collection *mongo.Collection
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection: database.Collection("auctions"),
	}
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	//  Inicia a rotina de fechamento automático
	// v1 com sleep go ar.waitAndCloseAuction(auctionEntity.Id)
	// v2 com channel depois de ver a solução com channel
	go ar.waitAndCloseWithChannel(ctx, auctionEntity.Id)

	return nil
}

// adicionado contexto para o gracefull shutdown par aque a goroutine nao fique travada
func (ar *AuctionRepository) waitAndCloseWithChannel(ctx context.Context, auctionId string) {
	auctionDuration := getAuctionDuration()

	expirationChan := time.After(auctionDuration) // time.After retorna um canal que receberá o tempo atual após a duração
	select {
	case <-expirationChan:
		ar.closeAuction(ctx, auctionId) // o canal recebeu o sinal de tempo esgotado
	case <-ctx.Done():
		logger.Info(fmt.Sprintf("Closure routine for auction %s cancelled due to context done", auctionId)) // Caso o contexto global seja cancelado (ex: shutdown do servidor)
		return
	}
}

func (ar *AuctionRepository) closeAuction(ctx context.Context, auctionId string) {
	filter := bson.M{"_id": auctionId, "status": auction_entity.Active}
	update := bson.M{"$set": bson.M{"status": auction_entity.Completed}}

	_, err := ar.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error(fmt.Sprintf("Error trying to close auction %s", auctionId), err)
		return
	}

	logger.Info(fmt.Sprintf("Auction %s successfully closed via channel after expiration", auctionId))
}

// v1 não bloqueio porém não permite cancelar
// func (ar *AuctionRepository) waitAndCloseAuctionWithSleep(auctionId string) {
// 	auctionDuration := getAuctionDuration()

// 	// Aguarda o tempo definido na env
// 	time.Sleep(auctionDuration)

// 	ctx := context.Background()
// 	filter := bson.M{"_id": auctionId, "status": auction_entity.Active}
// 	update := bson.M{"$set": bson.M{"status": auction_entity.Completed}}

// 	_, err := ar.Collection.UpdateOne(ctx, filter, update)
// 	if err != nil {
// 		logger.Error(fmt.Sprintf("Error trying to close auction %s", auctionId), err)
// 		return
// 	}

// 	logger.Info(fmt.Sprintf("Auction %s closed automatically after %s", auctionId, auctionDuration))
// }

func getAuctionDuration() time.Duration {
	durationStr := os.Getenv("AUCTION_DURATION")
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return time.Minute * 5 // Default caso erro ou vazio
	}
	return duration
}
