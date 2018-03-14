define([
	'app'
], function(app) {
	app.registerService('user', function($http, $rootScope) {
		var self = this;

		self.getUser = function(cb) {
			$http.get('/profile').then(function(response) {
				$rootScope.user = self.user = response.data.user;
				
				cb();
			});
		}
	});
});