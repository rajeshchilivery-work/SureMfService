package service

import (
	"SureMFService/database/cloudsql/repository"
	"SureMFService/database/firebase"
	"SureMFService/structs"
	"context"
	"fmt"
	"log"
)

// allowedATIs defines loan account type IDs eligible for EMI ROI delta comparison.
var allowedATIs = map[int]bool{2: true, 3: true, 4: true}

func GetEMIROIDelta(uid string) ([]structs.EMIROIDeltaItem, error) {
	// 1. Get user from DB to obtain user ID
	user, err := repository.GetSureUserByUID(uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 2. Get credit score from credit_details
	credit, err := repository.GetCreditDetailsByUserID(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credit details: %w", err)
	}

	// 3. Get retail account from Firebase creditData
	var retailAccount structs.FirebaseRetailAccount
	ctx := context.Background()
	doc, fbErr := firebase.FirestoreClient.Collection("creditData").Doc(uid).Get(ctx)
	if fbErr != nil {
		return nil, fmt.Errorf("firebase error for uid %s: %w", uid, fbErr)
	}
	if err := doc.DataTo(&retailAccount); err != nil {
		return nil, fmt.Errorf("failed to parse credit data: %w", err)
	}
	log.Printf("[emi-roi-delta] uid=%s, score=%d, loans=%d", uid, credit.Score, len(retailAccount.LN))

	// 4. Filter loans with ATI in {2, 3, 4} and build response
	var results []structs.EMIROIDeltaItem
	for _, loan := range retailAccount.LN {
		if !allowedATIs[loan.ATI] {
			continue
		}

		// 5. Get market rate for this loan type and credit score
		marketRate, err := repository.GetMarketRate(loan.ATI, credit.Score)
		if err != nil {
			log.Printf("skipping loan %s: %v", loan.ACC, err)
			continue
		}

		results = append(results, structs.EMIROIDeltaItem{
			ACC:  loan.ACC,
			OEMI: loan.EMI,
			OROI: loan.ROI,
			NROI: marketRate,
			NEMI: loan.SES.EMI,
		})
	}

	return results, nil
}
