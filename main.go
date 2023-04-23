package main

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

func main() {
	// Create a new AWS session as shown above
	// Replace with your desired AWS region
	region := "ap-south-1"
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})

	if err != nil {
		panic(err)
	}

	// Create an STS client using the session
	svc := sts.New(sess)

	// Get the caller identity
	resp, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})

	if err != nil {
		panic(err)
	}

	// Extract the account number from the ARN
	accountNumber := strings.Split(aws.StringValue(resp.Arn), ":")[4]

	fmt.Println("AWS account number:", accountNumber)
}
