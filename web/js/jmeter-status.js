/**
 * config & init chart
 */
window.dataPath = "./data/";
window.dataURL = window.dataPath + "today.json" + "?" + Date.parse(new Date());
window.chart;

/**
 * initialize chart
 * @param string containerId 
 * @param string dataContainId 
 * @param object data 
 */
function initStatusChart(containerId, dataContainId, data){
    if(window.chart != null || window.chart != undefined){
        window.chart.dispose();
    }
    window.chart = echarts.init(document.getElementById(containerId));
    window.chart.on("click",function(params){
        // load API data
        var file = data.files[params.dataIndex];
        if (file == undefined || file == null) return;
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
                    margin: 20
                }
            },
            {
                name : 'Error Rate',
                type : 'value',
                axisLabel : {
                    margin : 20,
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
 * initialize data for chart
 * @param {*} website 
 */
function initData(website, statusType){
    var times=[],errRates=[],counts=[],errCounts=[],files=[];

    $("#status-container,#apis-container").html("");
    // init status chart
    var data;
    if (statusType == "today" || statusType == undefined || statusType == null) {
        statusType = "today";
        data = website.data;
    } else {
        data = website.days;
    }
    $(data).each(function(k,v){
        times.push(statusType == "today" ? UTC2Local(v.executed_at) : v.date);
        counts.push(v.count);
        errCounts.push(v.err_count);
        errRates.push(v.err_rate);
        files.push(v.file);
    });

    initStatusChart('status-container', 'apis-container', {
        "times":times,
        "counts":counts,
        "errRates":errRates,
        "errCounts":errCounts,
        "files":files
    });
}

/**
 * init all
 */
function init(statusType){
    var websites;
    var defaultWebsite = localStorage.getItem("website");
    if(defaultWebsite == null) defaultWebsite=0;

    switch(statusType){
        case "week":
            window.dataURL = window.dataPath + "week.json" + "?" + Date.parse(new Date());
            break;
        case "month":
            window.dataURL = window.dataPath + "month.json" + "?" + Date.parse(new Date());
            break;
        case "today":
        default:
            window.dataURL = window.dataPath + "today.json" + "?" + Date.parse(new Date());
            break;
    }

    $.getJSON(window.dataURL, function(re){
        if(re != null && re.websites!=undefined && re.websites.length > 0){
            websites = re.websites;
            var defaultNum = defaultWebsite ? defaultWebsite : re.default;
            var website = websites[defaultNum];
            // init websites select
            $("#websites").html("");
            $(websites).each(function(k,v){
                $("#websites").append('<option value="'+ k +'">'+ v.name +'</option>');
            });
            $("#websites").unbind();
            $("#websites").change(function(){
                defaultWebsite = $(this).val();
                localStorage.setItem("website", defaultWebsite);
                initData(websites[defaultWebsite], statusType);
            });
            $("#websites").val(defaultNum);
            initData(website, statusType);
        }
    });
}

/**
 * timestamp to date string
 * @param string timestamp 
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

/**
 * UTC2Local
 * @param string UTCDateString 
 */
function UTC2Local(UTCDateString) {
    if(!UTCDateString) {
      return '-';
    }
    function formatFunc(str) {
      return str > 9 ? str : '0' + str
    }
    var date2 = new Date(UTCDateString);
    var year = date2.getFullYear();
    var mon = formatFunc(date2.getMonth() + 1);
    var day = formatFunc(date2.getDate());
    var hour = date2.getHours();
    hour = formatFunc(hour);
    var min = formatFunc(date2.getMinutes());
    var dateStr = mon+'/'+day+' '+hour+':'+min;
    return dateStr;
}

/**
 * tranfer html to entities
 * @param string str 
 */
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