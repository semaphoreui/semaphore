define([
	'app'
], function(app) {
	app.registerService('users', function($http, $rootScope) {
		var self = this;

		self.getUsers = function(cb) {
			$http.get('/users').then(function(response) {
				self.users = response.data;

				cb();
			});
		}
	});
});