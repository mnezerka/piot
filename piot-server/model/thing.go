package model

import (
    "go.mongodb.org/mongo-driver/bson/primitive"
)

const THING_TYPE_DEVICE = "device"
const THING_TYPE_SENSOR = "sensor"

const THING_CLASS_TEMPERATURE = "temperature"
const THING_CLASS_HUMIDITY = "humidity"
const THING_CLASS_PRESSURE = "pressure"

// Represents any device or app
type Thing struct {
    Id          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    Name        string `json:"name" bson:"name"`
    Alias       string `json:"alias" bson:"alias"`
    Type        string `json:"type" bson:"type"`
    Enabled     bool   `json:"enabled" bson:"enabled"`
    Created     int32  `json:"created" bson:"created"`
    OrgId       primitive.ObjectID `json:"org_id" bson:"org_id"`
    Org         *Org `json:"org" bson:"org"`
    Available   bool   `json:"available" bson:"available"`
    LastSeen    int32  `json:"last_seen" bson:"last_seen"`

    // The MQTT topic subscribed to receive thing availability
    AvailabilityTopic   string `json:"availability_topic" bson:"availability_topic"`
    AvailabilityYes     string `json:"availability_yes" bson:"availability_yes"`
    AvailabilityNo      string `json:"availability_no" bson:"availability_no"`

    //////////// Sensor data

    // The unit of measurement that the sensor is expressed in.
    Sensor SensorData `json:"sensor bson:"sensor""`
}

type SensorData struct {
    // The MQTT topic subscribed to receive sensor values.
    MeasurementTopic string `json:"measurement_topic" bson:"measurement_topic"`

    // Time when last measurement was received
    MeasurementLast int32 `json:"measurement_last" bson:"measurement_last"`

    // Type of sensor measurement
    Class string `json:"class" bson:"class"`

    // Defines the number of seconds sinc last measurement for which the
    // measurement is valid
    Validity int32  `json:"validity" bson:"validity"`

    // The unit of measurement that the sensor is expressed in.
    Unit string `json:"unit bson"unit""`
}
