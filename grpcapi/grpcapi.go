package grpcapi

import (
	"context"
	"database/sql"
	"log"
	"mailing/mdb"
	pb "mailing/proto"
	"net"
	"time"

	"google.golang.org/grpc"
)

type MailServer struct {
	pb.UnimplementedMailingListServiceServer
	db *sql.DB
}

func pbEntryToMdbEntry(pbEntry *pb.EmailEntry) mdb.EmailEntry {
	t := time.Unix(pbEntry.ConfirmedAt, 0)
	return mdb.EmailEntry{
		Id:          pbEntry.Id,
		Email:       pbEntry.Email,
		ConfirmedAt: &t,
		OptOut:      pbEntry.OptOut,
	}
}

func mdbEntryToPbEntry(mdbEntry *mdb.EmailEntry) pb.EmailEntry {
	return pb.EmailEntry{
		Id:          mdbEntry.Id,
		Email:       mdbEntry.Email,
		ConfirmedAt: mdbEntry.ConfirmedAt.Unix(),
		OptOut:      mdbEntry.OptOut,
	}
}

func emailResponse(db *sql.DB, email string) (*pb.EmailReponse, error) {
	entry, err := mdb.GetEmail(db, email)
	if err != nil {
		log.Println(err)
		return &pb.EmailReponse{}, err
	}
	if entry == nil {
		return &pb.EmailReponse{}, nil
	}
	res := mdbEntryToPbEntry(entry)

	return &pb.EmailReponse{EmailEntry: &res}, nil
}

func (s *MailServer) GetEmail(ctx context.Context, req *pb.GetEmailRequest) (*pb.EmailReponse, error) {
	log.Printf("grpc getEmail: %v]\n", req)
	return emailResponse(s.db, req.EmailAddr)
}

func (s *MailServer) GetEmailBatch(ctx context.Context, req *pb.GetEmailBatchRequest) (*pb.GetEmailBatchResponse, error) {
	log.Printf("grpc getbatchEmail: %v]\n", req)
	params := mdb.GetEmailBatchQueryParams{
		Page:  int(req.Page),
		Count: int(req.Count),
	}
	mdbEntries, err := mdb.GetEmailBatch(s.db, params)
	if err != nil {
		return &pb.GetEmailBatchResponse{}, err
	}

	pbEntries := make([]*pb.EmailEntry, 0, len(mdbEntries))

	for i := 0; i < len(mdbEntries); i++ {
		entry := mdbEntryToPbEntry(&mdbEntries[i])
		pbEntries = append(pbEntries, &entry)
	}
	return &pb.GetEmailBatchResponse{EmailEntry: pbEntries}, nil
}

func (s *MailServer) CreateEmail(ctx context.Context, req *pb.CreateEmailRequest) (*pb.EmailReponse, error) {
	log.Printf("grpc create Email: %v]\n", req)
	err := mdb.CreateEmail(s.db, req.EmailAddr)
	if err != nil {
		return &pb.EmailReponse{}, err
	}
	return emailResponse(s.db, req.EmailAddr)
}

func (s *MailServer) UpdateEmail(ctx context.Context, req *pb.UpdateEmailRequest) (*pb.EmailReponse, error) {
	log.Printf("grpc update Email: %v]\n", req)
	entry := pbEntryToMdbEntry(req.EmailEntry)
	err := mdb.UpdateEmail(s.db, entry)
	if err != nil {
		return &pb.EmailReponse{}, err
	}
	return emailResponse(s.db, entry.Email)
}

func (s *MailServer) DeleteEmail(ctx context.Context, req *pb.DeleteEmailRequest) (*pb.EmailReponse, error) {
	log.Printf("grpc update Email: %v]\n", req)
	err := mdb.Delete(s.db, req.EmailAddr)
	if err != nil {
		return &pb.EmailReponse{}, err
	}
	return emailResponse(s.db, req.EmailAddr)
}

func Serve(db *sql.DB, bind string) {
	listerner, err := net.Listen("tcp", bind)
	if err != nil {
		log.Fatalf("gRPC server error: failiour to bind %v\n", bind)
	}
	grpcServer := grpc.NewServer()
	MailServer := MailServer{db: db}
	pb.RegisterMailingListServiceServer(grpcServer, &MailServer)
	log.Printf("gRPC API server listenng on %v\n", bind)
	if err := grpcServer.Serve(listerner); err != nil {
		log.Fatal("gRPC server error: %V\n", err)
	}
}
