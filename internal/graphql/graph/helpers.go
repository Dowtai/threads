package graph

import (
	"threads/internal/entity"
)

func MapPostToGraphQL(post *entity.Post) *Post {
	if post == nil {
		return nil
	}

	return &Post{
		ID:              post.ID,
		Author:          post.Author,
		Title:           post.Title,
		Content:         post.Content,
		CommentsAllowed: post.CommentsAllowed,
	}
}

func MapCommentToGraphQL(comment *entity.Comment) *Comment {
	if comment == nil {
		return nil
	}

	return &Comment{
		ID:     comment.ID,
		Author: comment.Author,
		Text:   comment.Text,
	}
}

func MapPostListToGraphQL(postList []*entity.Post) []*Post {
	if postList == nil {
		return nil
	}

	posts := make([]*Post, len(postList))
	for i, post := range postList {
		posts[i] = MapPostToGraphQL(post)
	}

	return posts
}

func MapCommentListToGraphQL(commentList []*entity.Comment) []*Comment {
	if commentList == nil {
		return nil
	}

	comments := make([]*Comment, len(commentList))
	for i, comment := range commentList {
		comments[i] = MapCommentToGraphQL(comment)
	}

	return comments
}
