package main

import (
	"bytes"
	"fmt"
	"hake/hake"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

func main() {
	hs := hake.NewServer(
		hake.WithRegion("ap-south-1"),
	)

	var topic hake.Topic = "test_topic"

	_, err := hs.CreateTopic(topic)
	if err != nil {
		log.Fatalln("err", err)
	}

	err = hs.CreateSubscriberQueue(topic, "queue_1")
	if err != nil {
		log.Fatalln("err1", err)
	}

	err = hs.CreateSubscriberQueue(topic, "queue_1")
	if err != nil {
		log.Fatalln("err2", err)
	}

	err = hs.CreateSubscriberQueue(topic, "queue_2")
	if err != nil {
		log.Fatalln("err3", err)
	}

	id, err := hs.SendMessageOnTopic(topic, bytes.NewBuffer([]byte("hello world")))
	if err != nil {
		log.Fatalln("err while publish", err)
	}
	fmt.Println("messageId", id)

	sess, _ := session.NewSession(aws.NewConfig().WithRegion("ap-south-1"))
	client := sns.New(sess)
	out, err := client.ListSubscriptionsByTopic(
		&sns.ListSubscriptionsByTopicInput{
			TopicArn: aws.String(hs.Arn("sns", topic.String())),
		},
	)
	if err != nil {
		log.Fatalln("err while fetching subs", err)
	}
	fmt.Println("Listing subscriptions")
	for _, sub := range out.Subscriptions {
		fmt.Println(*sub.SubscriptionArn)
	}

}
