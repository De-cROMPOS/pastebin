package hashgen

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	pb "hashgenerator/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type HgProtoServer struct {
	pb.UnimplementedHasherServer
	DB *gorm.DB
}

func generateHash(text string, length int) (string, error) {

	if len(text) == 0 {
		return "", fmt.Errorf("text is empty")
	}

	// Хэшируем текст с солью
	hasher := sha256.New()
	salt := fmt.Sprintf("%v", time.Now().UnixNano())
	hasher.Write([]byte(text + salt))
	hashWithSalt := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	return hashWithSalt[:length], nil
}

func (s *HgProtoServer) GetHash(ctx context.Context, req *pb.HashRequest) (*pb.HashResponse, error) {
	hash, err := generateHash(req.GetText(), 10)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	exists, err := HashChecker(s.DB, hash)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	if exists {
		return nil, status.Errorf(codes.Aborted, "You're unlucky, your hash's being using")
	}

	return &pb.HashResponse{
		Hash: hash,
	}, nil
}

// func GetNewHash(text string) string {
// 	hash, err := generateHash(text, 10)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return hash
// }
