package services

// // si la route doit être protégée par OTP
// func isProtectedRoute(path string) bool {
// 	protectedRoutes := []string{
// 		"/v1/auth/sensitive-action",
// 		"/v1/auth/transfer-funds",
// 		"/v1/auth/delete-account",
// 	}

// 	return slices.Contains(protectedRoutes, path)
// }

// func RequireOTPValidation(next http.Handler, store *db.DBStore) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		//  si la route est protégée
// 		if !isProtectedRoute(r.URL.Path) {
// 			next.ServeHTTP(w, r)
// 			return
// 		}

// 		//  l'ID utilisateur depuis le JWT
// 		userToken := r.Header.Get("Authorization")
// 		claims, err := auth.VerifyJWT(userToken)
// 		if err != nil {
// 			log.Printf("Tentative d'accès sans JWT valide: %v", err)
// 			http.Error(w, "Unauthorized: Token invalide", http.StatusUnauthorized)
// 			return
// 		}

// 		//  l'email depuis Redis
// 		ctx := context.Background()
// 		email, err := RedisClient.Get(ctx, fmt.Sprintf("user_email:%s", claims.UserID)).Result()
// 		if err != nil {
// 			log.Printf("Email non trouvé pour userID %s, vérification en base de données...", claims.UserID)

// 			// chercher dans PostgreSQL si absent de Redis
// 			email, err = store.GetUserEmailByID(claims.UserID)
// 			if err != nil {
// 				log.Printf("Impossible de récupérer l'email pour userID %s", claims.UserID)
// 				http.Error(w, "Unauthorized: Impossible de vérifier l'utilisateur", http.StatusUnauthorized)
// 				return
// 			}

// 			// met l'email en cache Redis (5h)
// 			RedisClient.Set(ctx, fmt.Sprintf("user_email:%s", claims.UserID), email, 5*time.Hour)
// 		}

// 		//  si l'OTP a été validé pour cet email
// 		otpVerified, err := RedisClient.Get(ctx, fmt.Sprintf("otp_verified:%s", email)).Result()
// 		if err != nil || otpVerified != "true" {
// 			log.Printf("Accès interdit : OTP non validé pour %s", email)
// 			http.Error(w, "Unauthorized: OTP requis", http.StatusUnauthorized)
// 			return
// 		}

// 		log.Printf("Accès autorisé pour %s (OTP validé)", email)
// 		next.ServeHTTP(w, r)
// 	})
// }
