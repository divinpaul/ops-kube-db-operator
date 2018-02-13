package core

func CreateDatabaseIfNotExists(req CreateDatabaseRequest) (*Database, error) {
	describe := Get().DescribeDBInstance
	create := Get().CreateDBInstance

	if db, err := describe(req.Name); err == nil && db != nil {
		return db, nil
	}

	db, err := create(req)

	if err != nil {
		return nil, err
	}

	return db, nil
}
