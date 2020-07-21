package util

import (
	"log"
	"os"

	"github.com/Iwark/pushnotification"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/joho/godotenv"
)

func SendNotifications(fcmToken string, notificationBody string) {
	////Send message...
	//msg := &fcm.Message{
	//	To: fcmToken,
	//	Data: map[string]interface{}{
	//		"title": "Vepa",
	//		"body":  &notificationBody,
	//	},
	//}
	//// Create a FCM client to send the message.
	//// env.GoDotEnvVariable("FCM_SERVER_KEY")
	//client, err := fcm.NewClient(GoDotEnvVariable("FCM_SERVER_KEY"))
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//// Send the message and receive the response without retries.
	//response, err := client.Send(msg)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//log.Printf("%#v\n", response)
	//fmt.Println("notification sent...")
	//return

	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	push := pushnotification.Service{
		AWSAccessKey:         os.Getenv("AWSAccessKey"),
		AWSAccessSecret:      os.Getenv("AWSAccessSecret"),
		AWSSNSApplicationARN: os.Getenv("AWSSNSApplicationARN"),
		AWSRegion:            os.Getenv("AWSRegion"),
	}

	err := push.Send(os.Getenv("DeviceToken"), &pushnotification.Data{
		Alert: aws.String(notificationBody),
		Sound: aws.String("default"),
		Badge: aws.Int(1),
	})
	if err != nil {
		log.Fatal(err)
	}

}
