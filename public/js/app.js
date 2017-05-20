var app = angular.module('semaphore', ['scs.couch-potato', 'ui.router', 'ui.bootstrap', 'angular-loading-bar']);

couchPotato.configureApp(app);

app.config(['$httpProvider', function ($httpProvider) {
	$httpProvider.interceptors.push(['$q', '$injector', '$log', function ($q, $injector, $log) {
		return {
			request: function (request) {
				var url = request.url;
				if (url.indexOf('/tpl/') !== -1) {
					request.url = url = url.replace('/tpl/', '/public/html/');
				}

				if (!(url.indexOf('/public') !== -1 || url.indexOf('://') !== -1 || url.indexOf('uib/template') !== -1)) {
					request.url = "/api" + request.url;
					request.headers['Cache-Control'] = 'no-cache';
				}

				if (request.url.startsWith('/api') || request.url.startsWith('/public')) {
					request.url = request.url.substr(1);
				}

				return request || $q.when(request);
			}
		};
	}]);
}]);

app.run(['$rootScope', '$window', '$couchPotato', '$injector', '$state', '$http', function ($rootScope, $window, $couchPotato, $injector, $state, $http) {
	app.lazy = $couchPotato;

	$rootScope.$on('$stateChangeStart', function (event, toState, toParams, fromState, fromParams) {
		if (toState.pageTitle) {
			$rootScope.pageTitle = "Loading " + toState.pageTitle;
		} else {
			$rootScope.pageTitle = "Loading..";
		}
	});

	$rootScope.$on('$stateChangeSuccess', function (event, toState, toParams, fromState, fromParams) {
		$rootScope.previousState = {
			name: fromState.name,
			params: fromParams
		}

		if (toState.pageTitle) {
			$rootScope.pageTitle = toState.pageTitle;
		} else {
			$rootScope.pageTitle = "Ansible-Semaphore Page";
		}
	});

	$rootScope.refreshUser = function () {
		$rootScope.user = null;
		$rootScope.loggedIn = false;

		$rootScope.ws = null;

		$http.get('/user')
			.then(function (user) {
				$rootScope.user = user.data;
				$rootScope.loggedIn = true;

				$rootScope.refreshInfo();
				$rootScope.startWS();
			}, function () {
				$state.go('auth.login');
			});
	}

	$rootScope.startWS = function () {
		var ws_base = 'ws' + document.baseURI.substr(4);

		$rootScope.ws = new WebSocket(ws_base + 'api/ws');
		$rootScope.ws.onclose = function () {
			console.log('WS closed, retrying');
			setTimeout($rootScope.startWS, 2000);
		}

		$rootScope.ws.onmessage = function (e) {
			try {
				var d = JSON.parse(e.data);
				setTimeout(function () {
					$rootScope.$broadcast('task.' + d.type, d);
				}, 3000);
			} catch (_) { }
		}
	}

	$rootScope.refreshInfo = function (cb) {
		if (typeof cb != 'function') cb = function () { }

		$http.get('/info').success(function (info) {
			$rootScope.semaphore = info;
			cb();
		}).error(function () {
			cb(true);
		});
	}

	$rootScope.refreshUser();
}]);