package service_test

import (
    "testing"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson"
    "piot-server/test"
    "piot-server/model"
    "piot-server/service"
)

func TestPacketDeviceReg(t *testing.T) {
    const DEVICE = "device01"
    const SENSOR = "SensorAddr"

    ctx := test.CreateTestContext()

    test.CleanDb(t, ctx)

    s := ctx.Value("piotdevices").(*service.PiotDevices)

    // process packet for unknown device
    var packet model.PiotDevicePacket
    packet.Device = DEVICE

    var reading model.PiotSensorReading
    var temp float32 = 4.5
    reading.Address = SENSOR
    reading.Temperature = &temp
    packet.Readings = append(packet.Readings, reading)

    err := s.ProcessPacket(ctx, packet)
    test.Ok(t, err)

    // Check if defice is registered
    db := ctx.Value("db").(*mongo.Database)
    var thing model.Thing
    err = db.Collection("things").FindOne(ctx, bson.M{"name": DEVICE}).Decode(&thing)
    test.Ok(t, err)
    test.Equals(t, DEVICE, thing.Name)
    test.Equals(t, model.THING_TYPE_DEVICE, thing.Type)
    test.Equals(t, "available", thing.AvailabilityTopic)

    err = db.Collection("things").FindOne(ctx, bson.M{"name": SENSOR}).Decode(&thing)
    test.Ok(t, err)
    test.Equals(t, SENSOR, thing.Name)
    test.Equals(t, model.THING_TYPE_SENSOR, thing.Type)
    test.Equals(t, "temperature", thing.Sensor.Class)
    test.Equals(t, "value", thing.Sensor.MeasurementTopic)
}

func TestPacketDeviceDataUnassigned(t *testing.T) {

    const DEVICE = "device01"

    ctx := test.CreateTestContext()

    test.CleanDb(t, ctx)
    test.CreateThing(t, ctx, DEVICE)

    s := ctx.Value("piotdevices").(*service.PiotDevices)

    // process packet for know device
    var packet model.PiotDevicePacket
    packet.Device = DEVICE

    err := s.ProcessPacket(ctx, packet)
    test.Ok(t, err)

    // check if mqtt was called
    mqtt := ctx.Value("mqtt").(*service.MqttMock)
    test.Equals(t, 0, len(mqtt.Calls))
}

func TestPacketDeviceDataAssigned(t *testing.T) {

    const DEVICE = "device01"

    ctx := test.CreateTestContext()

    test.CleanDb(t, ctx)
    test.CreateThing(t, ctx, DEVICE)
    orgId := test.CreateOrg(t, ctx, "org1")
    test.AddOrgThing(t, ctx, orgId, DEVICE)

    s := ctx.Value("piotdevices").(*service.PiotDevices)

    // process packet for know device
    var packet model.PiotDevicePacket
    packet.Device = DEVICE
    ssid := "SSID"
    packet.WifiSSID = &ssid

    err := s.ProcessPacket(ctx, packet)
    test.Ok(t, err)

    // check if mqtt was called
    mqtt := ctx.Value("mqtt").(*service.MqttMock)
    test.Equals(t, 2, len(mqtt.Calls))

    test.Equals(t, "available", mqtt.Calls[0].Topic)
    test.Equals(t, "yes", mqtt.Calls[0].Value)

    test.Equals(t, "net/wifi/ssid", mqtt.Calls[1].Topic)
    test.Equals(t, "SSID", mqtt.Calls[1].Value)
}

func TestPacketDeviceReadingTempUnassigned(t *testing.T) {

    const DEVICE = "device01"

    ctx := test.CreateTestContext()

    test.CleanDb(t, ctx)
    test.CreateThing(t, ctx, DEVICE)

    s := ctx.Value("piotdevices").(*service.PiotDevices)

    // process packet for know device
    var temp float32 = 4.5
    var reading model.PiotSensorReading
    reading.Address = "SensorAddr"
    reading.Temperature = &temp

    var packet model.PiotDevicePacket
    packet.Device = DEVICE
    packet.Readings = append(packet.Readings, reading)

    err := s.ProcessPacket(ctx, packet)
    test.Ok(t, err)

    // check if mqtt was called
    mqtt := ctx.Value("mqtt").(*service.MqttMock)
    test.Equals(t, 0, len(mqtt.Calls))
}


func TestPacketDeviceReadingTempAssigned(t *testing.T) {

    const DEVICE = "device01"
    const SENSOR = "SensorAddr"

    ctx := test.CreateTestContext()

    test.CleanDb(t, ctx)
    test.CreateThing(t, ctx, DEVICE)
    test.CreateThing(t, ctx, SENSOR)
    orgId := test.CreateOrg(t, ctx, "org1")
    test.AddOrgThing(t, ctx, orgId, DEVICE)
    test.AddOrgThing(t, ctx, orgId, SENSOR)

    s := ctx.Value("piotdevices").(*service.PiotDevices)

    // process packet for know device
    var temp float32 = 4.5
    //var pressure float32 = 20.8
    //var humidity float32 = 95.5
    var reading model.PiotSensorReading
    reading.Address = SENSOR
    reading.Temperature = &temp
    //reading.Pressure= &pressure
    //reading.Humidity = &humidity

    var packet model.PiotDevicePacket
    packet.Device = DEVICE
    packet.Readings = append(packet.Readings, reading)

    err := s.ProcessPacket(ctx, packet)
    test.Ok(t, err)

    // check if mqtt was called
    mqtt := ctx.Value("mqtt").(*service.MqttMock)
    test.Equals(t, 4, len(mqtt.Calls))

    test.Equals(t, "available", mqtt.Calls[0].Topic)
    test.Equals(t, "yes", mqtt.Calls[0].Value)
    test.Equals(t, DEVICE, mqtt.Calls[0].Thing.Name)

    test.Equals(t, "available", mqtt.Calls[1].Topic)
    test.Equals(t, "yes", mqtt.Calls[1].Value)
    test.Equals(t, "SensorAddr", mqtt.Calls[1].Thing.Name)

    test.Equals(t, "value", mqtt.Calls[2].Topic)
    test.Equals(t, "4.5", mqtt.Calls[2].Value)
    test.Equals(t, "SensorAddr", mqtt.Calls[2].Thing.Name)

    test.Equals(t, "value/unit", mqtt.Calls[3].Topic)
    test.Equals(t, "C", mqtt.Calls[3].Value)
    test.Equals(t, "SensorAddr", mqtt.Calls[3].Thing.Name)
}



// Test Denial Of Service protection
func TestDOS(t *testing.T) {
    ctx := test.CreateTestContext()

    test.CleanDb(t, ctx)

    s := ctx.Value("piotdevices").(*service.PiotDevices)

    var packet model.PiotDevicePacket

    // send first packet
    packet.Device = "device01"
    err := s.ProcessPacket(ctx, packet)
    test.Ok(t, err)

    // check that sending same packet in short time frame is not possible
    err = s.ProcessPacket(ctx, packet)
    test.Assert(t, err != nil, "DOS protection doesn't work")

    // check that sending packet for different device is possible
    packet.Device = "device02"
    err = s.ProcessPacket(ctx, packet)
    test.Ok(t, err)
}
