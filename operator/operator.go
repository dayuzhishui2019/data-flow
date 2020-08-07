package operator

import (
	_ "dyzs/data-flow/operator/e_1400server"
	_ "dyzs/data-flow/operator/e_kafkaconsumer"
	_ "dyzs/data-flow/operator/e_onvif"
	_ "dyzs/data-flow/operator/h_1400client"
	_ "dyzs/data-flow/operator/h_1400digest"
	_ "dyzs/data-flow/operator/h_1400filter"
	_ "dyzs/data-flow/operator/h_1400tokafkamsg"
	_ "dyzs/data-flow/operator/h_datatowebsocket"
	_ "dyzs/data-flow/operator/h_downloadimage"
	_ "dyzs/data-flow/operator/h_kafkamsgto1400"
	_ "dyzs/data-flow/operator/h_kafkaproducer"
	_ "dyzs/data-flow/operator/h_toalidatahub"
	_ "dyzs/data-flow/operator/h_toalioss"
	_ "dyzs/data-flow/operator/h_uploadimage"
)
