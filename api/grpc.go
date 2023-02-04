package api

import (
	"context"
	"database/sql"
	"log"
	"net"
	"time"

	mdb "github.com/naufalihsan/mailer/db"
	pb "github.com/naufalihsan/mailer/proto"
	"google.golang.org/grpc"
)

type MailServer struct {
	pb.UnimplementedMailerServiceServer
	db *sql.DB
}

func ServeRPC(db *sql.DB, bind string) {
	listener, err := net.Listen("tcp", bind)

	if err != nil {
		log.Fatalf("gRPC server error: failure to bind")
	}

	grpcServer := grpc.NewServer()
	mailServer := MailServer{db: db}

	pb.RegisterMailerServiceServer(grpcServer, &mailServer)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("gRPC server error: %v\n", err)
	}
}

func (s *MailServer) GetEmail(ctx context.Context, req *pb.GetEmailRequest) (*pb.EmailResponse, error) {
	return emailResponse(s.db, req.Email)
}

func (s *MailServer) GetEmailBatch(ctx context.Context, req *pb.GetEmailBatchRequest) (*pb.EmailBatchResponse, error) {
	params := mdb.BatchEmailQueryParams{
		Page:  int(req.Page),
		Count: int(req.Count),
	}

	entryEmails, err := mdb.GetEmailBatch(s.db, params)

	if err != nil {
		return &pb.EmailBatchResponse{}, nil
	}

	protoEmails := make([]*pb.EmailEntry, 0, len(entryEmails))
	for i := 0; i < len(entryEmails); i++ {
		protoEmail := convertEntry(&entryEmails[i])
		protoEmails = append(protoEmails, &protoEmail)
	}

	return &pb.EmailBatchResponse{EmailEntry: protoEmails}, nil
}

func (s *MailServer) CreateEmail(ctx context.Context, req *pb.CreateEmailRequest) (*pb.EmailResponse, error) {
	err := mdb.InsertEmail(s.db, req.Email)

	if err != nil {
		return &pb.EmailResponse{}, err
	}

	return emailResponse(s.db, req.Email)
}

func (s *MailServer) UpdateEmail(ctx context.Context, req *pb.UpdateEmailRequest) (*pb.EmailResponse, error) {

	entry := convertProto(req.EmailEntry)

	err := mdb.UpdateEmail(s.db, entry)

	if err != nil {
		return &pb.EmailResponse{}, err
	}

	return emailResponse(s.db, entry.Email)
}

func (s *MailServer) DeleteEmail(ctx context.Context, req *pb.DeleteEmailRequest) (*pb.EmailResponse, error) {
	err := mdb.DeleteEmail(s.db, req.Email)

	if err != nil {
		return &pb.EmailResponse{}, err
	}

	return emailResponse(s.db, req.Email)
}

func emailResponse(db *sql.DB, email string) (*pb.EmailResponse, error) {
	entry, err := mdb.GetEmail(db, email)

	if err != nil {
		return &pb.EmailResponse{}, err
	}

	if entry == nil {
		return &pb.EmailResponse{}, nil
	}

	res := convertEntry(entry)

	return &pb.EmailResponse{EmailEntry: &res}, nil
}

func convertProto(pb *pb.EmailEntry) mdb.EmailEntry {
	unixTime := time.Unix(pb.ConfirmedAt, 0)
	return mdb.EmailEntry{
		Id:          pb.Id,
		Email:       pb.Email,
		ConfirmedAt: &unixTime,
		OptOut:      pb.OptOut,
	}
}

func convertEntry(db *mdb.EmailEntry) pb.EmailEntry {
	return pb.EmailEntry{
		Id:          db.Id,
		Email:       db.Email,
		ConfirmedAt: db.ConfirmedAt.Unix(),
		OptOut:      db.OptOut,
	}
}
