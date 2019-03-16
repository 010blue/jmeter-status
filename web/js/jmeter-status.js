/**
 * config & init chart
 * @param {*} containerId 
 * @param {*} data 
 */
window.dataPath = "./data/";
window.chart;

function initStatusChart(containerId, dataContainId, data){
    if(window.chart != null || window.chart != undefined){
        window.chart.dispose();
    }
    window.chart = echarts.init(document.getElementById(containerId));
    window.chart.on("click",function(params){
        // load API data
        var file = data.files[params.dataIndex];
        // use papaparse to get and parse data
        Papa.parse(window.dataPath + file, {
            download: true,
            complete: function(fileData) {
                var rows = fileData.data;
                // only show some column
                /**
                 * timeStamp 0
                 * elapsed 1
                 * label 2
                 * responseCode 3
                 * success 7
                 * failureMessage 8
                 */
                var html = '<table class="table table-sm table-hover tablesorter-blue">';
                // head
                html += '<thead><tr>';
                html += '<th>label</th>';
                html += '<th>timeStamp</th>';
                html += '<th>elapsed</th>';
                html += '<th>responseCode</th>';
                html += '<th>success</th>';
                html += '<th>failureMessage</th>';
                html += '</tr></thead>';
                html += '<tbody>';
                $(rows).each(function(rowK, columns){
                    if(rowK > 0 && columns != null && columns.length > 8) {
                        var label = columns[2];
                        var timeStamp = columns[0];
                        var success = columns[7];
                        if(rowK>0) timeStamp = timestamp2date(timeStamp);
                        html += '<tr>';
                        html += '<td>' + htmlEntities(label) +'</td>';
                        html += '<td>' + htmlEntities(timeStamp) +'</td>';
                        html += '<td>' + htmlEntities(columns[1]) +' ms</td>';
                        html += '<td>' + htmlEntities(columns[3]) +'</td>';
                        html += '<td' +(success=='false' ? ' class="text-danger font-weight-bold"' : '')+ '>' + htmlEntities(success) +'</td>';
                        html += '<td>' + htmlEntities(columns[8]) +'</td>';
                        html += '</tr>';
                    }
                });
                html += '</tbody></table>';

                $("#" + dataContainId).html(html);
                $("#" + dataContainId+" table").tablesorter();
            }
        });
    });
    var option= {
        title : {
            text: 'API Status',
            subtext: 'JMeter'
        },
        tooltip : {
            trigger: 'axis',
            formatter: 'Error: {c1}<br />Count: {c0}<br />ErrRate: {c2}%'
        },
        legend: {
            data:['Error', 'Count', 'Error Rate']
        },
        toolbox: {
            show : true,
            feature : {
                saveAsImage : {show: true},
                dataView : {readOnly: false}
            }
        },
        xAxis : [
            {
                type : 'category',
                data : data.times,
                axisLabel: {
                    rotate: 45
                }
            }
        ],
        yAxis : [
            {
                name : 'Count',
                type : 'value',
                axisLabel : {
                    margin: 30
                }
            },
            {
                name : 'Error Rate',
                type : 'value',
                axisLabel : {
                    margin : 30,
                    formatter: '{value} %'
                }
            }
        ],
        series : [
            {
                name: 'Count',
                type: 'bar',
                barGap: '-100%',
                itemStyle: {
                    normal: {
                        color: '#dcdcdc'
                    }
                },
                data: data.counts
            },
            {
                name: 'Error',
                type: 'bar',
                itemStyle: {
                    normal: {
                        color: '#dc3545'
                    }
                },
                data: data.errCounts
            },
            {
                name: 'Error Rate',
                type: 'line',
                data: data.errRates,
                yAxisIndex: 1,
                itemStyle: {
                    normal: {
                        color: '#fd7e14'
                    }
                },
                markPoint : {
                    data : [
                        {type : 'max', name: 'Max'},
                        {type : 'min', name: 'Min'}
                    ]
                },
                markLine : {
                    data : [
                        {type : 'average', name: 'Average'}
                    ]
                }
            }
        ]
    }
    chart.setOption(option);
}

/**
 * timestamp to date string
 * @param {*} timestamp 
 */
function timestamp2date(timestamp){
    var datetime = new Date();
    datetime.setTime(timestamp);
    var year = datetime.getFullYear();
    var month = datetime.getMonth() + 1 < 10 ? "0" + (datetime.getMonth() + 1) : datetime.getMonth() + 1;
    var date = datetime.getDate() < 10 ? "0" + datetime.getDate() : datetime.getDate();
    var hour = datetime.getHours()< 10 ? "0" + datetime.getHours() : datetime.getHours();
    var minute = datetime.getMinutes()< 10 ? "0" + datetime.getMinutes() : datetime.getMinutes();
    var second = datetime.getSeconds()< 10 ? "0" + datetime.getSeconds() : datetime.getSeconds();
    return month + "/" + date + " " + hour + ":" + minute;
}

function htmlEntities(str) {
    if(str == undefined || str == null || str == "") return "";
    var entitys = {
        '&' : '&amp;',
        '<' : '&lt;',
        '>' : '&gt;',
        '"' : '&quot;',
        "'" : '&apos;'
    };
    var regexp = new RegExp ('['+Object.keys(entitys).join('')+']','g');
    return str.replace(regexp,function(matched){
        return entitys[matched];
    });
} 