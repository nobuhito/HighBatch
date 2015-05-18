function getPanelHeading(d) {
    var titleStr = d["hostname"] + " / " + d["name"] + " / " + d["duration"].replace(/^(.*\d\.\d)\d+(s)$/, "$1$2");
    var exitClass = (d["exitCode"] > 0)?
            "text-danger glyphicon-warning-sign":
            "text-success glyphicon-ok";
    exitClass = (d["completed"] == "")?
        "text-success glyphicon glyphicon-repeat": exitClass;
    exitClass = (d["resolved"] != "")?
        "text-primary glyphicon glyphicon-thumbs-up": exitClass;
    var titleAlert = $("<div>").append(
        $("<span>").addClass(exitClass)
            .addClass("glyphicon")
            .attr({
                "aria-hidden": "true"
            }));
    titleStr = $(titleAlert).html() + " " + titleStr;
    var title = $("<a>").html(titleStr)
            .addClass("collapsed")
            .attr({
                "href": "#c-" + d["id"],
                "data-toggle": "collapse",
                "data-parent": "#accordion",
                "aria-expanded": "false",
                "aria-controls": "c-" + d["id"]
            });
    var panelTitle = $("<h4>").addClass("panel-title")
            .append(title);
    var panelHeading = $("<div>").addClass("panel-heading")
            .attr({
                "role": "tab",
                "id": "p-" + d["id"]
            })
            .append(panelTitle);
    return panelHeading;
}

function getTabPanel(d) {
    var tabPanel = $("<div>").attr("role", "tabpanel");
    var navTabs = $("<ul>").addClass("nav nav-tabs")
            .attr({
                "role": "tablist"
            });
    var tabContent = $("<div>").addClass("tab-content");
    var idActive = d["id"] + "output";
    var aActive = $("<a>").text("output")
            .attr({
                href: "#" + idActive,
                "aria-controls": idActive,
                role: "tab",
                "data-toggle": "tab"
            });
    var preesntationActive = $("<li>").addClass("active")
            .attr({
                role: "presentation"
            })
            .append(aActive)
            .appendTo(navTabs);
    var outputStr = d["output"];
    if (d["error"] != "") {
        var re = new RegExp("(" + d["error"] + ")", "gim");
        outputStr = d["output"].replace(re, function(match, p1, offset, string) {
            console.log(match);
            return "<span class=\"bg-danger text-danger\">" + match + "</span>";
        }).trim();
    }
    var output = $("<pre>").addClass("console").html(outputStr);
    var tabPaneActive = $("<div>").addClass("tab-pane active")
            .attr({
                id: idActive,
                role: "tabpanel"
            })
            .append(output);
    tabContent.append(tabPaneActive).appendTo(navTabs);

    for (i in d["assets"]) {
        var id = d["id"] + d["assets"][i].replace(/[^\w\d]/, "");
        var a = $("<a>").text(d["assets"][i])
                .attr({
                    href: "/source/" + d["name"] + "/" + d["assets"][i],
                    "data-target": "#" + id,
                    "aria-controls": id,
                    role: "tab",
                    "data-toggle": "tabajax"
                })
                .click(function(e) {
                    var $this = $(this),
                        loadurl = $this.attr("href"),
                        target = $this.attr("data-target");

                    $.ajax({
                        url: loadurl,
                        cache: true
                    }).done(function(data) {
                        var html = hljs.highlightAuto(data).value;
                        $(target).html("<pre class=\"hljs\">" + html + "</pre>");
                    });
                    $this.tab("show");
                    return false;
                });
        var preesntation = $("<li>").append(a)
                .attr("role", "presentation")
                .appendTo(navTabs);
        var asset = $("<pre>");
        var tabPane = $("<div>").addClass("tab-pane")
                .attr({
                    id: id,
                    role: "tabpanel"
                })
                .append(asset);
        tabContent.append(tabPane).appendTo(navTabs);
    }

    tabPanel.append(navTabs).append(tabContent);
    return tabPanel;
}

function getPanelCollapse(d) {
    var info = $("<dl>").addClass("dl-horizontal");

    for (j in d) {
        var s = d[j];
        var notPrintItems = ["output", "durationInt", "durationAve", "id", "key"];
        if (notPrintItems.indexOf(j) > -1 || s == "" || s == null ) { continue };
        if (j == "chain") {
            var arrow = (d["onErrorStop"] == "" || d["exitCode"] == 0)?
                    " -> ": " -> X ";
            if (s != null && s != "") {
                s = arrow + s.join(", ");
            } else {
                s = ""
            }
        }
        if (["completed", "started", "resolved"].indexOf(j) > -1) {
            s = dt2datetime(s);
        }
        if (j == "duration") {
            s = s.replace(/^(.*\d\.\d)\d+(s)$/, "$1$2")
                .replace(/([hms])/g, "$1 ").replace("m s", "ms");
        }
        if (j == "route") {
            if (s != null && s != "") {
                s = s.join(" -> ") + " ->";
            } else {
                s = "";
            }
        }
        if (j == "schedule") {
            if (s != "Manual" && s != "WebHook") {
                var jwd = ["日", "月", "火", "水", "木", "金", "土"];
                var c = s.split(" ");
                c.reverse();
                var wd = (/^[0-6]$/.test(c[0]))? jwd[c[0]]:c[0];
                s = wd + "曜日 " + c[1] + "月 " + c[2] + "日 ";
                s = s + c[3] + "時 " + c[4] + "分 " + c[5] + "秒";
            }
        }
        s = "<span style=\"font-family:monospace\">" + s + "</span>";
        $("<dt>").text(j.toUpperCase()).appendTo(info);
        $("<dd>").html(s).appendTo(info);
    }


    var panelBody = $("<div>").addClass("panel-body")
            .append(info);

    if (d["completed"] != "") {

        var tabPanel = getTabPanel(d);
        panelBody.append(tabPanel);

        if (d["schedule"] != "WebHook") {
            panelBody.append(
                $("<a>").addClass("execute btn btn-danger")
                    .attr("href", "/exec/" + d["key"])
                    .text("ReExecute")
                    .prepend(
                        $("<span>").addClass("glyphicon glyphicon-flash")
                            .attr("aria-hidden", "true")
                    )
            );
        }

        if (d["exitCode"] != 0 && d["resolved"] == "") {
            panelBody.append(
                $("<a>").addClass("resolve btn btn-primary")
                    .attr("href", "/resolve/" + d["id"])
                    .text("Resolve")
                    .prepend(
                        $("<span>").addClass("glyphicon glyphicon-thumbs-up")
                            .attr("aria-hidden", "true")
                    )
            );

        }
    } else {
        var progress = $("<div>").addClass("progress");
        var v = getProgresValue(d["durationAve"], d["started"]);
        var id = "progress-" + d["key"] + d["hostname"] + d["started"];
        var progressBar = $("<div>").addClass("progress-bar")
                .attr({
                    "id": id,
                    role: "progressbar",
                    "aria-valuenow": v,
                    "aria-valuemin": 0,
                    "aria-valuemax": 100,
                    "data-duration": d["durationAve"],
                    "data-started": d["started"]
                })
                .css({
                    width: v + "%"
                })
                .text(v + "%").appendTo(progress);

        panelBody.append(progress);
        setInterval(function(id) {
            progressBarUpdate(id);
        }, 1000, id);
    }
    var panelCollapse = $("<div>").addClass("panel-collapse collapse")
            .attr({
                "role": "tabpanel",
                "aria-labelledby": "p-" + d["id"],
                "id": "c-" + d["id"]
            })
            .append(panelBody);
    return panelCollapse;
}

function getProgresValue(durationAve, started) {
    var m = started.match(/^(\d{4})(\d{2})(\d{2})(\d{2})(\d{2})(\d{2})/);
    var starttime = new Date(m[1], m[2] - 1, m[3], m[4], m[5], m[6], 0);
    var duration = new Date() - starttime;
    var v = (duration * 100 / (durationAve / 1000000)).toString()
            .replace(/(.*\.\d).*/, "$1");
    return (v > 100)? 100: v;
}

function progressBarUpdate(target) {
    var t = $("#" + target);
    var value = getProgresValue(t.attr("data-duration"), t.attr("data-started"));
    t.attr("aria-valuenow", value)
        .css("width", value + "%")
        .text(value + "%");
}

function dt2datetime(dt) {
    var datetime = "";
    m = dt.match(/(\d{4})(\d{2})(\d{2})(\d{2})(\d{2})(\d{2})/);
    datetime = m[1] + "/" + m[2] + "/" + m[3] + " " + m[4] + ":" + m[5] + ":" + m[6];
    return datetime;
}

function getTree(items, cb) {
    $.ajax({
        url: "/worker/list",
        cache: false
    }).done( function(data) {
        var workers = {};
        for (var w in data) {
            console.log(data);
            workers[data[w]["host"]] = data[w]["isAlive"];
        }
        console.log(workers);
        var tree = [];
        for (var hostname in items) {
            alltag = 0;
            var h = {};
            var icon = (workers[hostname] == 1)? "glyphicon glyphicon-tasks": "glyphicon glyphicon-question-sign text-danger";
            h.text = hostname;
            h.icon = icon;
            h.href = "#" + hostname;
            h.nodes = [];
            for (var key in items[hostname]) {
                var k = {};
                k.text = items[hostname][key]["name"];
                k.icon = "glyphicon glyphicon-time";
                k.href = "#" + hostname + "/" + key;
                k.nodes = [];
                var tag = 0;
                for (var i in items[hostname][key]["error"]) {
                    error = items[hostname][key]["error"][i];
                    var item = {
                        "text": dt2datetime(error),
                        "icon": "glyphicon glyphicon-exclamation-sign",
                        "href": "#p-" + error
                    };
                    k.nodes.push(item);
                    tag++;
                }
                if (tag > 0) { k.tags = [tag]; }
                h.nodes.push(k);
                alltag = alltag + tag;
            }
            if (alltag > 0) { h.tags = [alltag]; }
            tree.push(h);
        }
        cb(tree);
    });
}

function doTreeview(tree) {
    $("#nav").treeview({
        data: tree,
        showBorder: false,
        showTags: true,
        collapseIcon: "glyphicon glyphicon-chevron-down",
        expandIcon: "glyphicon glyphicon-chevron-right",
        enableLinks: true,
        onNodeSelected: function(event, data) {
            if (/^#p\-/.test(data.href)) {
                $(data.href).children().children().click();
            } else {
                load("/data/" + data.href.replace("#", ""), false);
            }
        }
    });
}

function render(d) {
    var panelHeading = getPanelHeading(d);
    var panelCollapse = getPanelCollapse(d);
    var panel = $("<div>").addClass("panel panel-default")
            .append(panelHeading)
            .append(panelCollapse)
            .appendTo($("#accordion"));
}

function load(path, bothTree) {
    $("#article").html("");
    var panelGroup = $("<div>").addClass("panel-group")
            .attr({
                id: "accordion",
                role: "tablist",
                "aria-multiselectable": true
            })
            .appendTo($("#article"));

    $.ajax({
        url: path,
        cache: false
    }).done( function(data) {

        var durationTable = {};
        for (i in data) {
            var d = data[i];

            if (!(d["hostname"] + d["key"] in durationTable)) {
                durationTable[d["hostname"] + d["key"]] = {
                    count: 0,
                    value: 0
                };
            }
            var t = durationTable[d["hostname"] + d["key"]];
            if (d["durationInt"] != 0) {
                var  i = {
                    count: parseInt(t.count) + 1,
                    value: parseInt(t.value) + parseInt(d["durationInt"])
                };
                durationTable[d["hostname"] + d["key"]] = i;
            }
        }

        var items = {};
        for (i in data) {
            var d = data[i];
            var t = durationTable[d["hostname"] + d["key"]];
            if (t.count > 0) {
                d.durationAve = (t.value / t.count);
            }

            setTimeout(render, 100, d);

            if (bothTree == true) {
                var h = d["hostname"];
                if (!(h in items)) { items[h] = {}; }
                if (!(d["key"] in items[h])) { items[h][d["key"]] = {}; }
                items[h][d["key"]]["name"] = d["name"];

                var k = d["key"];
                if (!("error" in items[h][k])) { items[h][k]["error"] = []; }
                if (d["exitCode"] > 0 && d["resolved"] == "") {
                    items[h][k]["error"].push(d["id"]);
                }
            }
        }

        if (bothTree == true) {
            getTree(items, function(tree) {
                doTreeview(tree);
            });
        }

        $(".collapse").collapse('hide');
    });
}

var items = [
    {
        name: "Name",
        placeholder: "タスクの名前",
        type: "text",
        key: true

    },
    {
        name: "Description",
        placeholder: "タスクの詳細等",
        type: "textarea"
    },
    {
        name: "Cmd",
        placeholder: "実行するコマンド",
        type: "text"
    },
    {
        name: "Schedule",
        placeholder: "実行するスケジュール ( Cron方式で 秒 分 時 日 月 曜日 )",
        type: "text"
    },
    {
        name: "Chain",
        placeholder: "次に登録するタスク",
        type: "select",
        func: function() {
            $.ajax({
                url:"/tasks/data"
            }).done(function(data) {
                for (var i in data) {
                    $("#form_Chain").append($("<option>").text(i + ": " + data[i].name));
                }
            });
        }
    },
    {
        name: "Error",
        placeholder: "異常終了と判定する正規表現",
        type: "text"
    },
    {
        name: "OnErrorStop",
        type: "bool",
        label: "異常終了の時は次に進めずにストップする"
    },
    {
        name: "Assets",
        type: "file",
        label: "バッチファイルやSQLファイルなど"
    }
];

function getSpecInput() {

    var form = $("<form>").addClass("form-horizontal");
    for (var i in items) {
        var grp = $("<div>").addClass("form-group");
        var label = $("<label>").addClass("col-sm-3 control-label");
        label.attr("for", "form_" + items[i].name).text(items[i].name).appendTo(grp);
        var type = items[i].type;
        if (type == "text") {
            var input = $("<input>").addClass("form-control");
            input
                .attr({
                    type: items[i].type,
                    id: "form_" + items[i].name,
                    placeholder: items[i].placeholder
                });
            if (!items[i].key) {
                input.attr("disabled", "disabled");
            } else {
                setTimeout(function(id) {
                    $("#"+id).on("keyup", function() {
                        if ($("#"+id).val().length > 0) {
                            enableAll();
                        } else {
                            disableAll();
                        }
                    });
                }, 1000, "form_" + items[i].name);
            }
            input.appendTo($("<div>").addClass("col-sm-9").appendTo(grp));
        } else if (type == "textarea") {
            var textarea = $("<textarea>").addClass("form-control");
            textarea
                .attr({
                    rows: (items[i].rows || 1),
                    disabled: "disabled",
                    id: "form_" + items[i].name,
                    placeholder: items[i].placeholder
                })
                .appendTo($("<div>").addClass("col-sm-9").appendTo(grp));
        } else if (type == "select") {
            var select = $("<select>").addClass("form-control");
            select
                .attr({
                    id: "form_" + items[i].name,
                    disabled: "disabled"
                })
                .append($("<option>").append("選択▼ (複数指定する場合はspec.tomlを直接編集してください)"))
                .appendTo($("<div>").addClass("col-sm-9").appendTo(grp));
        } else if (type == "bool") {
        	  var check = $("<input>");
        	  check
        		    .attr({
        			      type: "checkbox",
                    disabled: "disabled",
        			      id: "form_" + items[i].name
        		    })
        		    .appendTo($("<div>").addClass("col-sm-9").appendTo(grp).css("padding-top", 7))
        		    .parent().append($("<span>").text(items[i].label).css("padding-left", "3px"));
        } else if (type== "file") {

            var drag_outer = $("<div>").addClass("col-sm-9");
            drag_outer
                .attr("id", "file_form_" + items[i].name)
                .appendTo(grp);
        }
        grp.appendTo(form);
    }

    form.append($("<hr>"));

    form.append(
        $("<div>").addClass("form-group")
            .append($("<div>").addClass("col-sm-offset-3 col-sm-9")
                    .append($("<button>").addClass("btn btn-primary")
                            .attr({
                                id: "addTaskButton",
                                type: "button",
                                "data-loading-text": "Send data..."
                            })
                            .text("Add tasks")
                           )
                   )
    );
    setTimeout(function() {
        $("#addTaskButton").on("click", function() {
            postTask();
        });
    }, 1000);

    // http://qiita.com/emadurandal/items/37fae542938907ef5d0c
    Function.prototype.toJSON = Function.prototype.toString;
    var jsonText = JSON.stringify(items);
    var parser = function(k,v){
        return v.toString().indexOf('function') === 0 ? eval('('+v+')') : v;
    };
    var funcJson = JSON.parse(jsonText, parser);
    for (var i in funcJson) {
        if (funcJson[i].func != undefined) {
            funcJson[i].func();
        }
    }

    return $("<div>").append(form);

}

function enableAll() {
    for (var i in items) {
        if (items[i].type == 'file') {
            if ($("#file_form_"+items[i].name).html() == "") {
                enableUplaod(items[i]);
            }
        } else {
            $("#form_" + items[i].name).prop("disabled", false);
        }
    }
}

function disableAll() {
    for (var i in items) {
        if (items[i].type == 'file') {
            disableUpload(items[i]);
        } else {
            if (!items[i].key) {
                $("#form_" + items[i].name).prop("disabled", true);
            }
        }
    }
}

function disableUpload(item) {
    $("#file_form_"+item.name).html("");
}

function enableUplaod(item) {
    console.log(item);

    var drag = $("<div>")
            .attr("id", "form_" + item.name);
    drag
        .addClass("drag")
        .append($("<p>")
                .text("アップロードするファイルをドロップ または")
               );

    var uploadBtn =$("<button>");
    uploadBtn
        .attr("id", "uploadBtn")
        .text("ファイルを選択")
        .change(function() {
            var files = this.files;
            addFiles(files);
        });
    var btnGrp = $("<span>").addClass("btn-group");
    btnGrp
        .append($("<input>")
                .attr({
                    type: "file",
                    name: "uploadfile"
                }).attr("multiple", "multiple"))
        .append(uploadBtn)
        .click(function() {
            $("#uploadBtn").click();
        })
        .appendTo(drag);
    drag.appendTo($("#file_form_"+item.name));
    setTimeout(function(item) {
        setupDragDrop(item);
    }, 100, item);
}

function setupDragDrop(item) {
    var elm = $("#form_" + item.name);
    elm
        .on("drop", function(e) {
            e.preventDefault();
            var files = e.originalEvent.dataTransfer.files;
            addFiles(files, item);
        })
        .on("dragenter", function(e) {
            e.stopPropagation();
            e.preventDefault();
            elm.addClass("bg-info");
        })
        .on("dragover", function(e) {
            e.stopPropagation();
            e.preventDefault();
        });
    $(document)
        .on("drop", function(e) {
            e.stopPropagation();
            e.preventDefault();
        })
        .on("dragenter", function(e) {
            e.stopPropagation();
            e.preventDefault();
        })
        .on("dragover", function(e) {
            e.stopPropagation();
            e.preventDefault();
            elm.removeClass("bg-info");
        });
}

var fd = new FormData();
function addFiles(files, item) {
    var elm = $("#form_" + item.name);
    var filesLen = files.length;
    for (var i = 0; i < filesLen; i++) {
        fd.append(item.name, files[i]);
        elm.parent().append($("<div>").text(files[i].name));
    }
    elm.removeClass("bg-info");
}

function postTask() {
    var btn =$("#addTaskButton").button("loading");
    for (var i in items) {
        var item = items[i];
        var id = "#form_" + item.name;
        if (item.type == 'file') { continue; }
        var data = $("#form_" + item.name).val();
        if (item.type == "bool") {
            data = ($(id).prop('checked'))? "on": "off";
        }
        fd.append(item.name, data);
    }

    $.ajax({
        url: "/task",
        type: "POST",
        data: fd,
        cache: false,
        processData: false,
        contentType: false
    }).done(function(data) {
        btn.button("reset");
    });
}

function task() {
    form = getSpecInput();
    $("#main").html(form.html());

}

function tasks() {
    $("#nav_tasks").addClass("active");
    $.ajax({
        url: "/tasks/data",
        cache: false
    }).done( function(data) {
        $("#main").html(jsonToTable(data));
    });
}

function conf() {
    $("#nav_conf").addClass("active");
    $.ajax({
        url: "/conf/data",
        cache: false
    }).done( function(data) {
        $("#main").html(jsonToTable(data, true));
    });
}

function workers() {
    $("#nav_workers").addClass("active");
    $.ajax({
        url: "/workers/data",
        cache: false
    }).done( function(data) {
        $("#main").html(jsonToTable(data, true));
    });
}

function jsonToTable(obj, isPrintNull) {
    if (isPrintNull == null || isPrintNull == undefined) isPrintNull = false;
    var table = $("<table>").addClass("table table-condensed");
    for (var i in obj) {
    	  if (!isPrintNull) {
            if (obj[i] == "" ||
                obj[i] == undefined ||
                (typeof obj[i] == "object" && Object.keys(obj[i]).length == undefined)) {
                continue;
            }
        }
        var tr = $("<tr>");
        tr.append($("<th>").text(i));
        if (typeof obj[i] == "object") {
            tr.append(jsonToTable(obj[i], isPrintNull));
        } else {
            tr.append($("<td>").text(obj[i]));
        }
        tr.appendTo(table);
    }
    return table;
}

function index() {
    load("/data", true);

    ("#nav").parent().hover(
        function() {
            $("#nav-wrap").removeClass().addClass("col-sm-6");
            $("#article-wrap").removeClass().addClass("col-sm-6");
        },
        function() {
            $("#nav-wrap").removeClass().addClass("col-sm-4");
            $("#article-wrap").removeClass().addClass("col-sm-8");
        }
    );

    $(function () {
        $('.panel-group').on('shown.bs.collapse', function (e) {
            var offset = $('.panel.panel-default > .panel-collapse.in').offset();
            if(offset) {
                $('html,body').animate({
                    scrollTop: $('.panel-collapse.in').siblings('.panel-heading').offset().top - 70
                }, 200);
            }
        });
    });
}