package main

import (
	"context"
	"fmt"
	"log"
	"net"

	blogpb "../proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//BlogServiceServer struct
type BlogServiceServer struct {
}

//CreateBlog function
func (s *BlogServiceServer) CreateBlog(ctx context.Context, req *blogpb.CreateBlogReq) (*blogpb.CreateBlogRes, error) {
	blog := req.GetBlog() // get protobuf from request
	data := BlogItem{     // converting into blogitem type to convert to bson
		AuthorID: blog.GetAuthourId(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
	}
	result, err := blogdb.InsertOne(mongoctx, data) // insert into database, insertone contains id
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}
	old := result.InsertedID.(primitive.ObjectID)
	blog.Id = old.Hex() //adding id to blog
	return &blogpb.CreateBlogRes{Blog: blog}, nil
}

//ReadBlog function
func (s *BlogServiceServer) ReadBlog(ctx context.Context, req *blogpb.ReadBlogReq) (*blogpb.ReadBlogRes, error) {
	old, err := primitive.ObjectIDFromHex(req.GetId()) // getting id from request and convert to mongodb id to retrieve data
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Could not convert to object ID: %v", err))
	}
	result := blogdb.FindOne(ctx, bson.M{"_id": old}) //finding with help of mongodb id
	data := BlogItem{}                                // to decode result and store
	if err := result.Decode(&data); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Could not find blog with object ID %s: %v", req.GetId(), err))
	}
	response := &blogpb.ReadBlogRes{
		Blog: &blogpb.Blog{
			Id:        old.Hex(),
			AuthourId: data.AuthorID,
			Title:     data.Title,
			Content:   data.Content,
		},
	}
	return response, nil
}

//UpdateBlog function
func (s *BlogServiceServer) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogReq) (*blogpb.UpdateBlogRes, error) {
	blog := req.GetBlog()                               // get blog data from request
	old, err := primitive.ObjectIDFromHex(blog.GetId()) // convert to mongodb id
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Could not convert to mongodb id: %v", err),
		)
	}
	update := bson.M{ // convert data to bson document
		"author_id": blog.GetAuthourId(),
		"title":     blog.GetTitle(),
		"content":   blog.GetContent(),
	}
	filter := bson.M{"_id": old}                                                                                            // convert to search by id
	result := blogdb.FindOneAndUpdate(ctx, filter, bson.M{"$set": update}, options.FindOneAndUpdate().SetReturnDocument(1)) //return updated document with context and filter
	decoded := BlogItem{}
	err = result.Decode(&decoded)
	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Could not find the document with ID: %v", err),
		)
	}
	return &blogpb.UpdateBlogRes{
		Blog: &blogpb.Blog{
			Id:        decoded.ID.Hex(),
			AuthourId: decoded.AuthorID,
			Title:     decoded.Title,
			Content:   decoded.Content,
		},
	}, nil
}

//DeleteBlog function
func (s *BlogServiceServer) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogReq) (*blogpb.DeleteBlogRes, error) {
	old, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Could not convert to Object Id: %v", err))
	}
	_, err = blogdb.DeleteOne(ctx, bson.M{"_id": old}) //returns result with struct of deleted ones
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Could not find and delete with the Id %s: %v", req.GetId(), err))
	}
	return &blogpb.DeleteBlogRes{
		Success: true,
	}, nil
}

//ListBlogs function
func (s *BlogServiceServer) ListBlogs(req *blogpb.ListBlogReq, stream blogpb.BlogService_ListBlogsServer) error {
	data := &BlogItem{} // blogitem to write decoded data
	cursor, err := blogdb.Find(context.Background(), bson.M{})
	if err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Unknown internal error: %v", err))
	}
	defer cursor.Close(context.Background()) //at the end it closes
	for cursor.Next(context.Background()) {  // returns boolean and loops to send all data to client
		err := cursor.Decode(data) // decode data at pointer and write to data
		if err != nil {
			return status.Errorf(codes.Unavailable, fmt.Sprintf("Could not decode data: %v", err))
		}
		stream.Send(&blogpb.ListBlogRes{ //sending blog through stream
			Blog: &blogpb.Blog{
				Id:        data.ID.Hex(),
				AuthourId: data.AuthorID,
				Content:   data.Content,
				Title:     data.Title,
			},
		})
	}
	if err := cursor.Err(); err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Unknown cursor error: %v", err))
	}
	return nil
}

//BlogItem struct, converted to bson, bson package uses meta information to assign keys
type BlogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

var db *mongo.Client
var blogdb *mongo.Collection
var mongoctx context.Context

func main() {
	fmt.Println("Starting server on port: 50051")
	listener, err := net.Listen("tcp", ":50051") //50051 default port of grpc, listening

	if err != nil {
		log.Fatalf("Unable to listen on port: 50051: %v", err)
	}

	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...) // create grpc server
	srv := &BlogServiceServer{}

	blogpb.RegisterBlogServiceServer(s, srv) //Register service with server

	fmt.Println("Connecting to Mongo database....")                                            // mongo client initialize
	mongoctx = context.Background()                                                            // context
	db, err := mongo.Connect(mongoctx, options.Client().ApplyURI("mongodb://localhost:27017")) // connect takes context and options
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping(mongoctx, nil) // checking if connection is success
	if err != nil {
		log.Fatalf("Could not connect to Mongodb: %v", err)
	} else {
		fmt.Println("Connected to Mongodb")
	}

	blogdb = db.Database("mydb").Collection("blog") // taking collection from databse to global variable to use in methods

	if err := s.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

	fmt.Println("Stopping the server..")
	s.Stop()
	listener.Close()
	fmt.Println("Closing database connection")
	db.Disconnect(mongoctx)
	fmt.Println("Over.")
}
