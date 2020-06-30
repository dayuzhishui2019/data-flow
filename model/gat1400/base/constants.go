package base

const CONTENT_TYPE  = "application/VIID+JSON"

const (
	REGIST_RETURN_AUTHORIZATION = "WWW-Authenticate"
	NONCE         = "nonce"
	QOP           = "qop"
	REALM         = "realm"
	OPAQUE        = "opaque"
	RESPONSE      = "response"
	ALGORITHM     = "algorithm"
	URI           = "uri"
	NC            = "nc"
	CNONCE        = "cnonce"
)

const (
	URL_REGIST = "/VIID/System/Register"

	URL_UNREGIST = "/VIID/System/UnRegister"

	URL_KEEPALIVE = "/VIID/System/Keepalive"

	URL_TIME = "/VIID/System/Time"

	URL_FACES = "/VIID/Faces"

	URL_PERSONS = "/VIID/Persons"

	URL_VEHICLE = "/VIID/MotorVehicles"

	URL_FILE = "/VIID/Files"

	URL_VIDEOSLICE = "/VIID/VideoSlices"

	URL_IMAGE = "/VIID/Images"

	URL_NOMOTORS = "/VIID/NonMotorVehicles"

	URL_ACOPENLOG = "/VIID/AcOpenLog"

	URL_ACOPENLOGIMG = "/VIID/AcOpenLogImg"

	URL_ACOPENLOGVIDEO = "/VIID/AcOpenLogVideo"

	URL_WIFIMODEL = "/VIID/WifiModel"

	URL_RFIDMODEL = "/VIID/RfidModel"

	URL_GPSMODEL = "/VIID/GpsModel"

	URL_APE = "/VIID/APEs"

	URL_DISPOSITIONNOTIFICATIONS = "/VIID/DispositionNotifications"

	AUTHORIZATION = "WWW-Authenticate"

	URL_USERIDENTIFY = "User-Identify"

	URL_CONTENTTYPE = "application/VIID+JSON"

	URL_SUBSCRIBENOTIFICATION = "/VIID/SubscribeNotifications"

	URL_REPORTCHANNELSINFO = "/VIID/api/dag/dispatch/v1/ReportChannelsInfo"

	URL_UPDATECHANNEL = "/VIID/api/dag/dispatch/v1/SubPlatformUpdateChannel"

	URL_WIFIDEVICE = "/VIID/WifiDevice"

	URL_WIFIDATA = "/VIID/WifiData"

	URL_HIGHALTDEVICE = "/VIID/HighAltDevice"

	URL_HIGHALTDEVICEPIC = "/VIID/HighAltDevicePic"

	URL_SENSORDEVICE = "/VIID/SensorDevice"

	URL_SENSORDEVICEALARM = "/VIID/SensorDeviceAlarm"

	URL_SENSOR = "/VIID/Sensor"

	URL_SENSORDATA = "/VIID/SensorData"

	URL_GEODEVICE = "/VIID/GeoDevice"

	URL_RECHARGEGARAGE = "/VIID/RechargeGarage"

	URL_RECHARGEGARAGEALARM = "/VIID/RechargeGarageAlarm"

	URL_MANHOLECOVER = "/VIID/ManholeCover"

	URL_MANHOLECOVERALARM = "/VIID/ManholeCoverAlarm"

	URL_FIREDEVICE = "/VIID/FireDevice"

	URL_FIREDEVICEALARM = "/VIID/FireDeviceAlarm"

	URL_ENVIRONMENTDETECT = "/VIID/EnvironmentDetect" // 环境检测

	URL_WATERQUALITYDETECT = "/VIID/WaterQulityDetect" // 水质检测

	URL_INTELLIGENTSTREETLAMP = "/VIID/IntelligentStreetLamp" // 智能路灯

	URL_GASDETECT = "/VIID/GasDetect" // 燃气检测

	URL_FIREDEVICE1 = "/VIID/FireDevice_1" // 1.烟感

	URL_FIREDEVICE2 = "/VIID/FireDevice_2" // 2.电表

	URL_FIREDEVICE3 = "/VIID/FireDevice_3" // 3.室外消火栓

	URL_FIREDEVICE4 = "/VIID/FireDevice_4" // 4. 电弧

	URL_FIREDEVICE5 = "/VIID/FireDevice_5" // 5.室内消火栓

	URL_FIREDEVICE6 = "/VIID/FireDevice_6" // 6.水泵

	URL_FIREDEVICE7 = "/VIID/FireDevice_7" // 7.压力表
)
