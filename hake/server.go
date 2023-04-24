package hake

import (
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sts"
)

const (
	Sns = "sns"
	Sqs = "sqs"
)

var accSingleton sync.Once

// Server is hake server side code
type Server struct {
	// nc is sns client to interact with aws sns
	nc *sns.SNS

	// qc is sqs client to interact with aws sns
	qc *sqs.SQS

	// tc is sts client to fetch account details for formulating ARNs
	tc *sts.STS

	region AwsRegion

	mu      sync.Mutex // for concurrent access to the vars below
	account AwsAccount

	accessKey AwsAccessKey
	secretKey AwsSecretKey
}

// creates a new hake server instance
func NewServer(opts ...ServerOption) *Server {
	// create clients
	s := &Server{region: "ap-south-1"}

	for _, opt := range opts {
		opt(s)
	}

	credsConfig := aws.NewConfig()
	if s.accessKey != "" && s.secretKey != "" {
		credsConfig = credsConfig.WithCredentials(
			credentials.NewStaticCredentials(
				s.accessKey.String(),
				s.secretKey.String(),
				"",
			),
		)
	}

	sess, err := session.NewSession(
		aws.NewConfig().WithRegion(string(s.region)),
		credsConfig,
	)
	s.secretKey = AwsSecretKey("") // remove from memory and let GC do its work

	if err != nil {
		panic(err)
	}

	s.nc = sns.New(sess)
	s.qc = sqs.New(sess)
	s.tc = sts.New(sess)

	return s
}

// options for initialising hake server
type ServerOption func(*Server)

// uses the region below if not then default is used
func WithRegion(region AwsRegion) ServerOption {
	return func(s *Server) {
		s.region = region
	}
}

// aws access key
func WithAccessKey(accessKey AwsAccessKey) ServerOption {
	return func(s *Server) {
		s.accessKey = accessKey
	}
}

// note that this is a temp access only, it is removed from memory afterwards
func WithSecretKey(secretKey AwsSecretKey) ServerOption {
	return func(s *Server) {
		s.secretKey = secretKey
	}
}

// Create a new kafka like topic
func (s *Server) CreateTopic(topic Topic) (topicARN string, err error) {

	// Todo: check if this is an idempotent operation or not ..
	out, err := s.nc.CreateTopic(&sns.CreateTopicInput{Name: aws.String(topic.String())})
	if err != nil {
		return "", err
	}

	return *out.TopicArn, nil

}

// Creates subscriber to a topic with correct policies
func (s *Server) CreateSubscriberQueue(topic Topic, queueName string) (err error) {
	// 1. create queue with correct policy
	_, err = s.qc.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
		Attributes: map[string]*string{
			"DelaySeconds":      aws.String("0"),
			"VisibilityTimeout": aws.String("120"), // 2min
			"Policy":            aws.String(s.Policy(topic, queueName)),
		},
	})

	if err != nil {
		return err
	}

	// 2. create subscription to the topic
	_, err = s.nc.Subscribe(&sns.SubscribeInput{
		Endpoint: aws.String(s.Arn(Sqs, queueName)),
		Protocol: aws.String(Sqs),
		TopicArn: aws.String(s.Arn(Sns, topic.String())),
		Attributes: map[string]*string{
			"RawMessageDelivery": aws.String("true"),
		},
	})
	if err != nil {
		log.Printf("err while creating subscription: %s\n", err)
		return
	}

	return nil

}

func (s *Server) SendMessageOnTopic(topic Topic, reader io.Reader) (string, error) {

	// create topic arn
	topicArn := s.Arn(Sns, topic.String())

	var buffer [10 * 1024 * 1024]byte // Todo: reduce the size of this buffer
	n, err := reader.Read(buffer[:])
	if err != nil {
		return "", err
	}

	out, err := s.nc.Publish(&sns.PublishInput{
		Message:  aws.String(string(buffer[:n])),
		TopicArn: &topicArn,
	})
	if err != nil {
		return "", err
	}

	return *out.MessageId, nil
}

// gets account number for current connection
func (s *Server) GetAccountNumber() (account string, _ error) {

	var resp *sts.GetCallerIdentityOutput
	var err error

	accSingleton.Do(func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		if s.account != "" {
			return
		}

		// Get the caller identity
		resp, err = s.tc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
		if err != nil {
			log.Printf("err while calling GetCallerIdentity: %s", err)
			return
		}

		// Extract the account number from the ARN
		identifier := strings.Split(aws.StringValue(resp.Arn), ":")
		if len(identifier) < 5 {
			err = fmt.Errorf("identifier too small")
			return
		}

		s.account = AwsAccount(identifier[4])
	})

	account = s.account.String()
	return account, err
}

// get complete Arn for a resource and resourceName
func (s *Server) Arn(resource string, resourceName string) string {

	s.GetAccountNumber() //populate account number first

	s.mu.Lock()
	defer s.mu.Unlock()

	return fmt.Sprintf(
		"arn:aws:%s:%s:%s:%s",
		resource,
		s.region,
		s.account,
		resourceName,
	)
}

func (s *Server) Policy(topic Topic, queue string) string {
	topicArn := s.Arn(Sns, topic.String())
	queueArn := s.Arn(Sqs, queue)

	policyString := fmt.Sprintf(
		`
		{
			"Version": "2012-10-17",
			"Statement": [
			  {
				"Sid": "AllowSNSPublishToQueue",
				"Effect": "Allow",
				"Principal": {
				  "AWS": "*"
				},
				"Action": "sqs:SendMessage",
				"Resource": "%s",
				"Condition": {
				  "ArnEquals": {
					"aws:SourceArn": "%s"
				  }
				}
			  }
			]
		  }
		`, queueArn, topicArn,
	)

	return policyString
}
