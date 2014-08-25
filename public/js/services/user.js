define([
	'app'
], function(app) {
	app.registerService('user', function($http, $rootScope) {
		var self = this;

		self.getUser = function(cb) {
			$http.get('/profile').success(function(data) {
				$rootScope.user = self.user = data.user;
				
				cb();
			});
		}
	});
});