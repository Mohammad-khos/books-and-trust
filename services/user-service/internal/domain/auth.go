package domain


type Authenticator interface {
    GenerateToken(userID string) (string, error)
    VerifyToken(token string) (string, error) 
}