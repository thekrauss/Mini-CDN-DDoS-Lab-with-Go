package services

// OTP 6 chiffres
// func generateOTP() string {
// 	rand.NewSource(time.Now().UnixNano())
// 	return fmt.Sprintf("%06d", rand.Intn(1000000))
// }

// func StoreOTP(email, otp string) error {
// 	ctx := context.Background()

// 	requestsKey := fmt.Sprintf("otp_requests:%s", email)
// 	count, _ := RedisClient.Get(ctx, requestsKey).Int()
// 	if count >= 3 {
// 		log.Printf("Trop de demandes d'OTP pour %s", email)
// 		return fmt.Errorf("trop de tentatives, réessayez plus tard")
// 	}

// 	err := RedisClient.Set(ctx, fmt.Sprintf("otp:%s", email), otp, 5*time.Minute).Err()
// 	if err != nil {
// 		log.Printf("Impossible de stocker l'OTP dans Redis: %v", err)
// 		return err
// 	}

// 	RedisClient.Incr(ctx, requestsKey)
// 	RedisClient.Expire(ctx, requestsKey, 10*time.Minute)

// 	log.Printf("OTP stocké pour %s", email)
// 	return nil
// }

// func SendOTPByEmail(email string) (string, error) {
// 	otp := generateOTP()

// 	// Stocke l'OTP en Redis
// 	err := StoreOTP(email, otp)
// 	if err != nil {
// 		return "", err
// 	}

// 	subject := "Code de récupération de mot de passe"
// 	body := fmt.Sprintf("Votre code de récupération est : %s\n\nCe code est valide pour 5 minutes.", otp)

// 	err = sendEmail(email, subject, body)
// 	if err != nil {
// 		log.Printf("Échec de l'envoi de l'email OTP: %v", err)
// 		return "", fmt.Errorf("échec de l'envoi de l'email OTP")
// 	}

// 	log.Printf("OTP envoyé avec succès à %s", email)
// 	return otp, nil
// }

// func VerifyOTP(email, inputOTP string) bool {
// 	ctx := context.Background()
// 	otpKey := fmt.Sprintf("otp:%s", email)

// 	storedOTP, err := RedisClient.Get(ctx, otpKey).Result()
// 	if err != nil {
// 		log.Printf("OTP non trouvé ou expiré pour %s", email)
// 		return false
// 	}

// 	if storedOTP == inputOTP {
// 		log.Printf("OTP valide pour %s", email)

// 		_, delErr := RedisClient.Del(ctx, otpKey).Result()
// 		if delErr != nil {
// 			log.Printf("Erreur lors de la suppression de l'OTP pour %s : %v", email, delErr)
// 		}
// 		return true
// 	}

// 	log.Printf("OTP incorrect pour %s", email)
// 	return false
// }
