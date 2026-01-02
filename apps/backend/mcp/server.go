package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewServer() *mcp.Server {
	return mcp.NewServer(&mcp.Implementation{
		Name:    "monarch-mcp",
		Version: "1.0.0",
	}, &mcp.ServerOptions{})
}