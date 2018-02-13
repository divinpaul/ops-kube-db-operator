package core

import "sync"

type CreateDatabaseRequest struct {
	Name string
}

type Database struct{}

type DescribeDBInstance func(name string) (*Database, error)
type CreateDBInstance func(req CreateDatabaseRequest) (*Database, error)
type GenerateRandomPassword func() string

type Functions struct {
	CreateDBInstance       CreateDBInstance
	DescribeDBInstance     DescribeDBInstance
	GenerateRandomPassword GenerateRandomPassword
}

var mutex = &sync.RWMutex{}
var functions Functions

func Set(funcs Functions) {
	mutex.Lock()
	functions = funcs
	mutex.Unlock()
}

func Get() Functions {
	mutex.RLock()
	defer mutex.RUnlock()
	return functions
}
