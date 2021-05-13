package databases

import (
"context"
"fmt"
	logger "github.com/zztroot/zztlog"
"go.mongodb.org/mongo-driver/mongo"
"go.mongodb.org/mongo-driver/mongo/options"
"zzt_blog/config"
)

type MongoConn struct {
	Config config.ConfigRead
	DB     *mongo.Database
	Conn   *mongo.Client
}

func (m *MongoConn) ConnectMongo() (*mongo.Database, *mongo.Client, error) {
	conf, err := m.Config.RConfigConn()
	if err != nil {
		logger.Error(err)
		return nil, nil, err
	}
	dbHost := conf.Get("mongodb.db_host")
	dbPort := conf.Get("mongodb.db_port")
	dbName := conf.GetString("mongodb.db_name")
	c := fmt.Sprintf(`mongodb://%s:%s`, dbHost, dbPort)
	clientOptions := options.Client().ApplyURI(c)
	m.Conn, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		logger.Error(err)
		return nil, nil, err
	}
	if err := m.Conn.Ping(context.TODO(), nil); err != nil {
		logger.Error(err)
		return nil, nil, err
	}
	m.DB = m.Conn.Database(dbName)
	return m.DB, m.Conn, nil
}
