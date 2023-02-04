package main

import (
	"context"
	"log"
	"time"

	"github.com/alexflint/go-arg"
	pb "github.com/naufalihsan/mailer/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func CreateEmail(client pb.MailerServiceClient, email string) *pb.EmailResponse {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.CreateEmail(ctx, &pb.CreateEmailRequest{Email: email})
	logResponse(res, err)

	return res
}

func GetEmail(client pb.MailerServiceClient, email string) *pb.EmailResponse {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.GetEmail(ctx, &pb.GetEmailRequest{Email: email})
	logResponse(res, err)

	return res
}

func GetEmailBatch(client pb.MailerServiceClient, count int32, page int32) *pb.EmailBatchResponse {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.GetEmailBatch(ctx, &pb.GetEmailBatchRequest{Page: page, Count: count})

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	for i := 0; i < len(res.EmailEntry); i++ {
		log.Printf("item [%v of %v]: %s\n", i+1, len(res.EmailEntry), res.EmailEntry[i])
	}

	return res
}

func UpdateEmail(client pb.MailerServiceClient, entry *pb.EmailEntry) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.UpdateEmail(ctx, &pb.UpdateEmailRequest{EmailEntry: entry})
	logResponse(res, err)
}

func DeleteEmail(client pb.MailerServiceClient, email string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.DeleteEmail(ctx, &pb.DeleteEmailRequest{Email: email})
	logResponse(res, err)
}

func logResponse(res *pb.EmailResponse, err error) {
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if res.EmailEntry == nil {
		log.Fatalln("email not found")
	} else {
		log.Printf("response: %v\n", res.EmailEntry.Email)
	}
}

var args struct {
	GrpcAddr string `arg:"env:MAILER_GRPC_ADDR"`
}

func main() {
	arg.MustParse(&args)

	if args.GrpcAddr == "" {
		args.GrpcAddr = ":8081"
	}

	conn, err := grpc.Dial(args.GrpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalln("client not connected")
	}

	defer conn.Close()

	client := pb.NewMailerServiceClient(conn)

	emailAddress := "naufal.ihsan@mail.com"

	newEmail := CreateEmail(client, emailAddress)
	newEmail.EmailEntry.ConfirmedAt = 10000

	UpdateEmail(client, newEmail.EmailEntry)
	GetEmail(client, emailAddress)
	GetEmailBatch(client, 1, 1)
	DeleteEmail(client, emailAddress)
}
