package operator

import (
	_ "sunset/data-stream/operator/e_1400server"
	_ "sunset/data-stream/operator/e_kafkaconsumer"
	_ "sunset/data-stream/operator/h_1400client"
	_ "sunset/data-stream/operator/h_1400digest"
	_ "sunset/data-stream/operator/h_1400filter"
	_ "sunset/data-stream/operator/h_1400tokafkamsg"
	_ "sunset/data-stream/operator/h_downloadimage"
	_ "sunset/data-stream/operator/h_kafkamsgto1400"
	_ "sunset/data-stream/operator/h_kafkaproducer"
	_ "sunset/data-stream/operator/h_uploadimage"
	_ "sunset/data-stream/operator/h_datatowebsocket"
)
