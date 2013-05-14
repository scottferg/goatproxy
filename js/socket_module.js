angular.module("SocketModule", []);
angular.module("SocketModule").factory("socket", function ($rootScope) {
	var sock = new WebSocket("ws://" + window.document.location.host + "/socket");
	return {
		onopen: function (callback) {
			sock.onopen = function(m) {
				var args = arguments;
				$rootScope.$apply(function() {
					if (callback) { callback.apply(sock, args); }
				});
			}
		},
		onmessage: function (callback) {
			sock.onmessage = function(m) {
				var args = arguments;
				$rootScope.$apply(function() {
					if (callback) { callback.apply(sock, args); }
				});
			}
		},
		onerror: function (callback) {
			sock.onerror = function(m) {
				var args = arguments;
				$rootScope.$apply(function() {
					if (callback) { callback.apply(sock, args); }
				});
			}
		},
		onclose: function (callback) {
			sock.onclose = function(m, callback) {
				var args = arguments;
				$rootScope.$apply(function() {
					if (callback) { callback.apply(sock, args); }
				});
			}
		},
		send: function(message) {
			sock.send(message);
		}
	};
});
