function graph() {
    renderGanttSlider();
    renderGanttChart();

    $("a[data-toggle=\"tab\"]").on("shown.bs.tab", function(e) {
        if ($(e.target).attr("href") == "#relation") {
            renderRelationTree(addNoRelationTasklist);
        }
    });

    $("a[data-toggle=\"tab\"]").on("shown.bs.tab", function(e) {
        if ($(e.target).attr("href") == "#execTime") {
            renderExecTime();
        }
    });
}

function renderExecTime() {
    $.ajax({
        url: "/data"
    }).done(function(data) {
        var a = {};
        for (var i in data) {
            var d = data[i];
            if (!a[d.hostname]) { a[d.hostname] = {}; }
            if (!a[d.hostname][d.name]) {a[d.hostname][d.name] = []; }
            a[d.hostname][d.name].push(Math.ceil(d.durationInt/1000000000));
        }

        var d = [];
        for (var hostname in a) {
            for (var name in a[hostname]) {
                var vals = a[hostname][name];
                var item = {
                    hostname: hostname,
                    name: name,
                    data: medianArray(vals, 10),
                    info: [
                        { key: "mean",   val: s2str(d3.mean(vals)) },
                        { key: "median", val: s2str(d3.median(vals)) },
                        { key: "max",    val: s2str(d3.max(vals)) },
                        { key: "min",    val: s2str(d3.min(vals)) }
                    ]
                };
                d.push(item);
            }
        }
        createExecTimeTable("#execTime-data", d);
    });
}

function s2str(second) {
    switch (true) {
    case (second > 60 * 60):
        var h = Math.floor(second / 3600);
        var m = Math.floor((second % 3600) / 60);
        return h + "h " + m + "m";
        break;
    case (second > 60):
        var m = Math.floor(second / 60);
        return m + "m";
        break;
    default:
        return Math.floor(second) + "s";
    }
}

function createExecTimeTable(elm, data) {
    d3.select(elm).text("");

    var w = $(elm).width();
    var h = ($(elm).height() == 0)? 500: $(elm).height();
    var svg = d3.select(elm).append("svg")
            .attr("width", w)
            .attr("height", h)
            .attr("transform", "translate(10, 10)");

    var header = svg.selectAll(".header")
            .data([
                { key: "Host", x: 0 },
                { key: "Task", x: 200 },
                { key: "ExecTime(<-New)", x: 500,anchor: "middle" },
                { key: "Mean", x: 700, anchor: "end" },
                { key: "Median", x: 770, anchor: "end" },
                { key: "Max", x: 840, anchor: "end" },
                { key: "Min", x: 910, anchor: "end" }
            ])
            .enter()
            .append("g")
            .attr("class", "header")
            .attr("transform", function(d, i) {
                return "translate(20, 30)";
            });

    header.append("text")
        .text(function(d) {
            return d.key;
        })
        .attr("text-anchor", function(d) {
            return d.anchor || "start";
        })
        .attr("y", 0)
        .attr("x", function(d) {
            return d.x;
        });

    var row = svg.selectAll(".row").data(data)
            .enter()
            .append("g")
            .attr("class", "row")
            .attr("transform", function(d, i) {
                return "translate(20, " + ((i * 30) + 70) + ")";
            });

    row.selectAll(".val")
        .data(function(d) {
            var m = d3.median(d.data);
            var val = [];
            d.data.forEach(function(v, k) {
                val.push({
                    key: k,
                    value: ((v == 0 && m == 0)? 1: (v == 0)? 0: (v / m) * 10),
                    scale: ((v == 0 && m == 0)? 1: (v == 0)? 0: (v / m))
                });
            });
            return val;
        })
        .enter()
        .append("circle")
        .attr("cx", function(d, i) {
            return (i * 30 + 380);
        })
        .attr("cy", -5)
        .attr("r", function(d) {
            return (d.value != "NaN")? d.value: 0;
        })
        .attr("opacity", "0.9")
        .attr("fill", function(d) {
            switch (true) {
            case (d.scale > 1.5):
                return "#D9534F";
                break;
            case (d.scale < 0.5):
                return "#EC971F";
                break;
            default:
                return "#337AB7";
                break;
            }
        });

    row.append("text")
        .text(function(d, i) {
            return (i != 0 && data[i - 1].hostname == d.hostname)? "": d.hostname;
        })
        .attr("x", 0);

    row.append("text")
        .text(function(d) {
            return d.name;
        })
        .attr("text-anchor", "end")
        .attr("x", 350);

    row.selectAll(".info")
        .data(function(d) {
            return d.info;
        })
        .enter()
        .append("text")
        .text(function(d) {
            return d.val;
        })
        .attr("text-anchor", "end")
        .attr("x", function(d, i) {
            return 700 + (i *70);
        });

    var carray = [[0, 40], [1000, 40]];
    var line = d3.svg.line()
        .x(function(d) { return d[0];})
        .y(function(d) { return d[1];});
    svg.append("path")
        .attr("d", line(carray))
        .attr("class", "headerline")
        .attr("stroke", "black")
        .attr("stroke-width", "1");

}

function medianArray(array, count) {
    if (array.length <= count) { return array; }

    var odd = array.length % count;
    var div = Math.floor(array.length / count);

    var col = 0;
    var colInCount = 1;

    var data = [];
    for (var i=0; i<array.length; i++) {
        if (!data[col]) { data[col] = []; }
        data[col].push(array[i]);
        if (colInCount < div + ((col < odd)? 1: 0)) {
            colInCount++;
        } else {
            col++;
            colInCount = 1;
        }
    }

    data.forEach(function(d, i) {
        data[i] = d3.median(d);
    });

    return data;
}

function addNoRelationTasklist() {
    d3.selectAll("#relation-alert").text("");
    var relationTask = [];
    var nodes = $("#relation-data").data("nodes");
    for (var i in nodes) {
        relationTask.push(nodes[i].name);
    }

    var noRelationTask = [];
    var tasks = $("#relation-data").data("tasks");
    for (var i in tasks) {
        if (relationTask.indexOf(tasks[i].name) == -1) {
            renderTaskMessage("「" + tasks[i].name + "」 が使用されていません。", "warning");
            noRelationTask.push(tasks[i].name);
        }
    }
}

function renderTaskMessage(message, level) {
    $("#relation-alert").prepend(
        $("<div>")
            .addClass("alert alert-" + level)
            .attr("role", "alert")
            .text(message)
            .prepend($("<span>").addClass("glyphicon glyphicon-question-sign"))
    );
}

function findChild(tasks, key, parents) {
    for (var i in tasks) {
        if (tasks[i].name == key) {
            var task = tasks[i];
            if (parents.indexOf(task.name) != -1) {
                renderTaskMessage("「" + tasks[i].name + "」 がループしています。", "danger");
                return {name: "LOOP"};
            } else {
                if (task.chain == null || task.chain.length == 0) {
                    return {name:task.name};
                } else if (task.chain[0] == "") {
                    return {name:task.name};
                } else {
                    var children = [];
                    for (var j in task.chain) {
                        parents.push(task.name);
                        var child = findChild(tasks, task.chain[j], parents);
                        if (child != undefined) {
                            children.push(child);
                        }
                    }
                    return {
                        name:task.name,
                        children: children
                    };
                }
                break;
            }
        }
    }
    return undefined;
}

function renderRelationTree(cb) {
    $.ajax({
        url: "/tasks/data"
    }).done(function(tasks) {
        var data = {
            name: "Schedule起動",
            children: []
        };
        for (var i in tasks) {
            var task = tasks[i];

            if (task.schedule != null && task.schedule != "") {
                if (task.chain == null || task.chain.length == 0 || task.chain[0] == "") {
                    data.children.push({
                        name: task.name
                    });
                } else {
                    var inLoop = false;
                    var children = [];
                    for (var j in task.chain) {
                        var c = findChild(tasks, task.chain[j], [task.name]);
                        if (c != undefined) {
                            children.push(c);
                        }
                    }
                    if (!inLoop) {
                        data.children.push({
                            name: task.name,
                            children: children
                        });
                    }
                }
            }
        }
        $("#relation-data").data("tasks", tasks);
        createRelationTree("#relation-data", data);
        cb();
    });
}

function createRelationTree(elm, data) {
    d3.select(elm).text("");
    var w = $(elm).width();
    var h = ($(elm).height() == 0)? 500: $(elm).height();
    var svg = d3.select(elm).append("svg")
        .attr("width", w)
        .attr("height", h)
        .attr("transform", "translate(10, 10)");

    var tree = d3.layout.tree().size([w-10, h-10]);

    var nodes = tree.nodes(data);
    nodes[0].y = 0;
    nodes[0].x = 0;
    nodes.sort(function(a, b) {
        if (a.y < b.y) return -1;
        if (a.y > b.y) return  1;
        if (a.x < b.x) return -1;
        if (a.x > b.x) return  1;
        return 1;
        // return (b.y + b.x > a.y + a.x)? -1: 1;
    });
    var span = h / nodes.length;
    var left = nodes[0].x;
    for (var i in nodes) {
        if (i != 0) {
            if (nodes[i].x < left) {
                left = nodes[i].x;
            }
            nodes[i].y = span * i;
        }
    }
    nodes[0].x = 50;
    nodes[0].y = 30;
    $(elm).data("nodes", nodes);
    var links = tree.links(nodes);
    $(elm).data("links", links);

    var node = svg.selectAll(".node").data(nodes)
        .enter()
        .append("g")
        .attr("class", "node")
        .attr("transform", function(d) {
            return "translate(" + d.x + "," + d.y +")";
        });

    node.append("circle")
        .attr("r", 4)
        .attr("fill", "steelblue");

    node.append("text")
        .text(function(d) { return d.name; })
        .attr("x", function(d) {
            return (d.name == "Schedule起動")? -40: 5;
        })
        .attr("y", function(d) {
            return (d.name == "Schedule起動")? -10: 5;
        });

    var diagonal = d3.svg.diagonal().projection(function(d) {
        return [d.x, d.y];
    });

    svg.selectAll(".link")
        .data(links)
        .enter()
        .append("path")
        .attr("class", "link")
        .attr("fill", "none")
        .attr("stroke", "silver")
        .attr("d", diagonal);
}

function renderGanttChart() {

    var tasks = [];
    var taskNames = [];

    var taskStatus = {
        "SUCCEEDED": "bar",
        "FAILED": "bar-failed"
    };

    $.ajax({
        url: "/data"
    }).done(function(data) {
        for (var i in data) {
            var d = data[i];
            var start = dt2datetime(d.started);
            var end = dt2datetime(d.completed);
            if (start == undefined || end == undefined) {
                continue;
            }
            var task = {
                startDate: Date.parse(start),
                endDate: Date.parse(end),
                taskName: d.name,
                desc: ["host: " + d.hostname, start, " - " + end].join("\n"),
                status: (d.exitCode == 0)? "SUCCEEDED": "FAILED"
            };
            tasks.push(task);

            if (taskNames.indexOf(d.name) < 0) {
                taskNames.push(d.name);
            }
        }

        taskNames.sort();

        tasks.sort(function(a, b) {
            return a.endDate - b.endDate;
        });
        tasks.sort(function(a, b) {
            return a.startDate - b.startDate;
        });

        var offsetHour = 6; //hr

        var format = "%H:%M";
        switch (true) {
        case (offsetHour > 24) :
            format = "%m/%d %H:%M";
            break;
        case (offsetHour > 168):
            format = "%m/%d";
            break;
        }

        var gantt = d3.gantt();
        var width = $("#main").width() - 30;
        var height = (taskNames.length + 1) * 35;
        gantt
            .taskTypes(taskNames)
            .taskStatus(taskStatus)
            .width(width)
            .height(height)
            .container("#history-data")
            .tickFormat(format)
            .timeDomain([d3.time.hour.offset(Date.now(), offsetHour * -1), Date.now()])
            .timeDomainMode("fixed");

        $("#history-data").data("gantt", gantt);
        $("#history-data").data("tasks", tasks);
        gantt(tasks);
    });

}

function renderGanttSlider() {
    $("#history-graphRange").show();
    var slider = $("#history-graphRange")
            .css("width", "100%")
            .slider({
                id: "range",
            range: true,
            value: [0, 6],
            ticks: [0, 6, 12, 24, 72, 168],
                formatter: function(value) {
                    var from = (value[0]%24 == 0)? value[0]/24 + "日": value[0] + "h";
                    var to   = (value[1]%24 == 0)? value[1]/24 + "日": value[1] + "h";
                    return "過去 " + from + " - " + to + " の範囲を表示";
                }
            })
            .on("slideStop", function() {
                var gantt = $("#history-data").data("gantt");
                var tasks = $("#history-data").data("tasks");

                var val = this.value.split(",");
                var from = d3.time.hour.offset(Date.now(), val[1] * -1);
                var to = d3.time.hour.offset(Date.now(), val[0] * -1);

                var format = "%H:%M";
                switch (true) {
                case (val[1] - val[0] > 24) :
                    format = "%m/%d %H:%M";
                    break;
                case (val[1] - val[0] > 168):
                    format = "%m/%d";
                    break;
                }

                gantt
                    .timeDomain([from, to])
                    .tickFormat(format)
                    .redraw(tasks);
            });
}
