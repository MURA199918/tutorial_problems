package main

import (
	"context"
	"fmt"
	"log"

	blogpb "../proto"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Welcome client..")
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure()) //connection to localhost webpage
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := blogpb.NewBlogServiceClient(conn)
	blog := &blogpb.Blog{
		AuthourId: "1",
		Title:     "Hello world",
		Content:   "First publish",
	}
	blog1 := &blogpb.Blog{
		Id:        "602a5d6b2738b05ae3c7bde2",
		AuthourId: "1",
		Title:     "Hello new world",
		Content:   "First publish",
	}
	//requestCreate := &blogpb.CreateBlogReq{object}
	requestRead := &blogpb.ReadBlogReq{Id: "602a5d6b2738b05ae3c7bde2"}
	//requestUpdate := &blogpb.UpdateBlogReq{message1: createMessage{"1", "1", "Hello world", "First publish"}}
	requestDelete := &blogpb.DeleteBlogReq{Id: "602a5d6b2738b05ae3c7bde2"}

	responseCreate, _ := client.CreateBlog(context.Background(), &blogpb.CreateBlogReq{Blog: blog})
	responseRead, _ := client.ReadBlog(context.Background(), requestRead)
	responseUpdate, _ := client.UpdateBlog(context.Background(), &blogpb.UpdateBlogReq{Blog: blog1})
	responseDelete, _ := client.DeleteBlog(context.Background(), requestDelete)

	fmt.Println("Create Authour ID Response => ", responseCreate.Blog.AuthourId)
	fmt.Println("Create Title Response => ", responseCreate.Blog.Title)
	fmt.Println("Created Content Response => ", responseCreate.Blog.Content)

	if responseRead != nil {
		fmt.Println("Read operation successful")
		fmt.Printf(responseRead.Blog.AuthourId, responseRead.Blog.Title, responseRead.Blog.Content)
	}

	if responseUpdate != nil {
		fmt.Println("Update operation successful")
		fmt.Printf(responseUpdate.Blog.AuthourId, responseUpdate.Blog.Title, responseUpdate.Blog.Content)
	}

	if responseDelete != nil {
		fmt.Println("Delete operation successful")
	}
}
