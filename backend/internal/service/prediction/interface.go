package prediction

import (
	predictionIDL "ai-eino-interview-agent/api/model/prediction"
	"context"
)

type PredictionService interface {
	Predict(ctx context.Context, req *predictionIDL.PredictRequest, userID uint) (*predictionIDL.PredictResponse, error)
	ListPredictions(ctx context.Context, req *predictionIDL.ListPredictionRequest, userID uint) (*predictionIDL.ListPredictionResponse, error)
	GetPredictionDetail(ctx context.Context, req *predictionIDL.GetPredictionDetailRequest, userID uint) (*predictionIDL.GetPredictionDetailResponse, error)
}
