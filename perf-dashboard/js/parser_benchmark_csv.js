// reference: https://blog.harveydelaney.com/parsing-a-csv-file-using-node-javascript/
// TODO: this should be a common file
var fs = require("fs");
var parse = require("csv-parse");

var csvFile = "../data/benchmark.csv";

class BenchmarkData {
    constructor(StartTime,ActualDuration,Labels,NumThreads,ActualQPS,p50,p90,p99,
                cpu_mili_avg_telemetry_mixer,
                cpu_mili_max_telemetry_mixer,
                mem_MB_max_telemetry_mixer,
                cpu_mili_avg_fortioserver_deployment_proxy,
                cpu_mili_max_fortioserver_deployment_proxy,
                mem_MB_max_fortioserver_deployment_proxy,
                cpu_mili_avg_ingressgateway_proxy,
                cpu_mili_max_ingressgateway_proxy,
                mem_MB_max_ingressgateway_proxy){
        this.StartTime = StartTime;
        this.ActualDuration = ActualDuration;
        this.Labels = Labels;
        this.NumThreads = NumThreads;
        this.ActualQPS = ActualQPS;
        this.p50 = p50;
        this.p90 = p90;
        this.p99 = p99;
        this.cpu_mili_avg_telemetry_mixer = cpu_mili_avg_telemetry_mixer;
        this.cpu_mili_max_telemetry_mixer = cpu_mili_max_telemetry_mixer;
        this.mem_MB_max_telemetry_mixer =  mem_MB_max_telemetry_mixer;
        this.cpu_mili_avg_ingressgateway_proxy = cpu_mili_avg_ingressgateway_proxy;
        this.cpu_mili_max_ingressgateway_proxy =  cpu_mili_max_ingressgateway_proxy;
        this.mem_MB_max_ingressgateway_proxy = mem_MB_max_ingressgateway_proxy;
    }
}

const processData = (err, data) => {
    if (err) {
        console.log(`An error was encountered: ${err}`);
        return;
    }

    data.shift(); // only required if csv has heading row

    const perfDataList = data.map(row => new BenchmarkData(...row));

    analyseUsers(perfDataList);
}

fs.createReadStream(csvFile)
    .pipe(parse({ delimiter: ',' }, processData));

const analyseUsers = userList => {
    const ageSum = userList.reduce((acc, val) => acc += val.NumThreads, 0);
    const averageAge = ageSum / userList.length;
    console.log(averageAge);
}

