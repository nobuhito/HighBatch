function graph() {
    renderGanttSlider();
    renderGanttChart();
    renderRelationTree();
}

function findParent(tasks, key) {
    var parent = [];
    for (var i in tasks) {
        var task = tasks[i];
        if (task.chain == null || task.chain.length == 0 || task.chain[0] == "") {
            continue;
        }
        for (var j in task.chain) {
            if (task.chain[j] == key) {
                parent.push(task.name);
            }
        }
    }
    return (parent.length > 0)? parent: undefined;
}

function findChild(tasks, key) {
    for (var i in tasks) {
        if (tasks[i].name == key) {
            var task = tasks[i];
            if (task.chain == null || task.chain.length == 0) {
                return {name:task.name};
                // return task;
            } else if (task.chain[0] == "") {
                return {name:task.name};
                // return task;
            } else {
                children = [];
                for (var j in task.chain) {
                    child = findChild(tasks, task.chain[j]);
                    if (child != undefined) {
                        children.push(child);
                    }
                }
                return {
                    name:task.name,
                    children: children
                };
                // return task;
            }
            break;
        }
    }
    return undefined;
}

function renderRelationTree() {
    $.ajax({
        url: "/tasks/data"
    }).done(function(tasks) {
        var data = {
            name: "Tasks",
            children: []
        };
        for (var i in tasks) {
            var task = tasks[i];

            task.parent = findParent(tasks, task.name);

            if (task.chain == null || task.chain.length == 0 || task.chain[0] == "") {
                continue;
            }
            var children = [];
            for (var j in task.chain) {
                var c = findChild(tasks, task.chain[j]);
                if (c != undefined) {
                    children.push(c);
                }
            }
            data.children.push({
                name: task.name,
                children: children
            });
            // task.child = child;
        }
        createRelationTree("#relation", data);
    });
}

function createRelationTree(elm, data) {
    var w = $(elm).width();
    var h = ($(elm).height() == 0)? 500: $(elm).height();
    var svg = d3.select(elm).append("svg")
        .attr("width", w)
        .attr("height", h)
        .attr("transform", "translate(10, 10)");

    var tree = d3.layout.tree().size([w-10, h-10]);

    var nodes = tree.nodes(data);
    var c = nodes.length;
    for (var i in nodes) {
        if (i != 0) {
            nodes[i].y = nodes[i].y - (((c - i) * 10) % 9) * 15;
        }
    }
    var links = tree.links(nodes);

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
        .attr("x", 5)
        .attr("y", function(d) {
            return (d.name == "Tasks")? 15: 5;
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
            .container("#gantt")
            .tickFormat(format)
            .timeDomain([d3.time.hour.offset(Date.now(), offsetHour * -1), Date.now()])
            .timeDomainMode("fixed");

        $("#gantt").data("gantt", gantt);
        $("#gantt").data("tasks", tasks);
        gantt(tasks);
    });

}

function renderGanttSlider() {
    $("#graphRange").show();
    var slider = $("#graphRange")
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
                var gantt = $("#gantt").data("gantt");
                var tasks = $("#gantt").data("tasks");

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
