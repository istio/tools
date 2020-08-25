// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"math/big"
	"os"

	"github.com/dgrijalva/jwt-go"
)

type Jwks struct {
	Keys []*Jwk `json:"keys"`
}

type Jwk struct {
	Kty string `json:"kty"`
	E   string `json:"e"`
	N   string `json:"n"`
}

func generateToken(policyData SecurityPolicy, privateKey *rsa.PrivateKey) (string, error) {
	issuer := fmt.Sprintf("issuer-%d", policyData.RequestAuthN.NumJwks)
	if policyData.RequestAuthN.TokenIssuer != "" {
		issuer = policyData.RequestAuthN.TokenIssuer
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": issuer,
		"sub": "subject",
	})
	if policyData.RequestAuthN.InvalidToken {
		newPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return "", err
		}
		privateKey = newPrivateKey
	}
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func generateJwks(privateKey *rsa.PrivateKey) (string, error) {
	jwks := &Jwks{
		Keys: []*Jwk{
			{
				E:   base64.URLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes()),
				N:   base64.URLEncoding.EncodeToString((*privateKey.PublicKey.N).Bytes()),
				Kty: "RSA",
			},
		},
	}

	jwksBytes, err := json.Marshal(jwks)
	if err != nil {
		return "", err
	}
	return string(jwksBytes), nil
}

func writeTokenIntoFile(token string, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	_, err = file.WriteString(fmt.Sprintf(`"Authorization":"Bearer %s"`, token))
	if err != nil {
		file.Close()
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}
