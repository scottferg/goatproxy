var myApp = angular.module("debug", ["SocketModule", 'ui.bootstrap', 'ngSanitize']);
myApp.controller("MainController", function ($scope, socket) {
	$scope.connection = {};
	$scope.log = [];
	$scope.connection.status = "Disconnected";
    $scope.proxy_status = "";

    $scope.oneAtATime = true;

	socket.onopen(function (m) {
		$scope.connection.status = "Connected";
	});

	socket.onclose(function (m) {
		$scope.connection.status = "Disconnected";
	});

	socket.onerror(function (m) {
		$scope.connection.status = "Disconnected";
	});

	socket.onmessage(function (m) {
        message = angular.fromJson(m.data)

        if (message.type == "connect-success") {
            $scope.proxy_status = message.payload;
        } else if (message.type = "proxy-update") {
            $scope.log.push(message.payload);
        }
	});

	$scope.hasRows = function() {
		return $scope.log.length != 0;
	}

	$scope.renderBody = function(e) {
        function syntaxHighlight(json) {
            if (typeof json != 'string') {
                 json = JSON.stringify(json, undefined, 2);
            }
            json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
            return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
                var cls = 'number';
                if (/^"/.test(match)) {
                    if (/:$/.test(match)) {
                        cls = 'key';
                    } else {
                        cls = 'string';
                    }
                } else if (/true|false/.test(match)) {
                    cls = 'boolean';
                } else if (/null/.test(match)) {
                    cls = 'null';
                }
                return '<span class="' + cls + '">' + match + '</span>';
            });
        }

        console.log(e.body);
        if (e.httpev.Header !== undefined && e.httpev.Header['Content-Type'] !== undefined) {
            if (e.httpev.Header['Content-Type'][0] == "application/json") {
                return syntaxHighlight(JSON.stringify(angular.fromJson(e.body), undefined, 2));
            } else if (e.httpev.Header['Content-Type'][0] == "text/html") {
                return e.body;
            } else if (e.httpev.Header['Content-Type'][0] == "text/plain") {
                return e.body;
            }
        }
	}

    $scope.eventHeader = function(e) {
        result = "";

        var request = null;
        if (e.Request === undefined) {
            request = e
            result += "<i class='icon-chevron-right'></i>&nbsp;&nbsp;";

            result += request.Method + " ";
            result += request.RequestURI;
        } else {
            result += "<i class='icon-chevron-left'></i>&nbsp;&nbsp;";

            result += e.Proto + " ";
            result += e.Status;
        }

        return result;
    }
});

myApp.filter('reverse', function() {
	return function(items) {
		return items.slice().reverse();
	};
});

myApp.filter('token', function() {
	return function(request) {
        if (request) {
            return '->';
        }
        
        return '<-';
	};
});
