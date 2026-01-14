package auction

import (
	"context"
	"fullcycle-auction_go/configuration/database/mongodb"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestAuctionAutoClose(t *testing.T) {
	// Arrange

	ctx := context.Background()
	if err := godotenv.Load("../../../../cmd/auction/.env"); err != nil {
		log.Fatal("Error trying to load env variables")
		return
	}

	database, _ := mongodb.NewMongoDBConnection(ctx)

	os.Setenv("AUCTION_DURATION", "2s")
	repo := NewAuctionRepository(database)

	auction := &auction_entity.Auction{
		Id:        "test-123",
		Status:    auction_entity.Active,
		Timestamp: time.Now(),
	}

	// Act
	err := repo.CreateAuction(context.Background(), auction)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, auction_entity.Active, auction.Status) // Verifica se está aberto inicialmente
	time.Sleep(3 * time.Second)                            // Aguarda o tempo da env + margem de segurança
	auction, err = repo.FindAuctionById(ctx, auction.Id)   // Busca novamente do banco
	assert.Nil(t, err)
	assert.Equal(t, auction_entity.Completed, auction.Status)
}
