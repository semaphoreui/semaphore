define([
	'app'
], function(app) {
	app.registerService('users', function($http, $rootScope) {
		var self = this;

		self.getUsers = function(cb) {
			$http.get('/users').success(function(data) {
				self.users = data;

				cb();
			});
		}
	});
});