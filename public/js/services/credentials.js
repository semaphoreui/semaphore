define([
	'app'
], function(app) {
	app.registerService('credentials', function($http, $rootScope) {
		var self = this;

		self.getCredentials = function(cb) {
			$http.get('/credentials').success(function(data) {
				self.credentials = data;
				
				cb();
			});
		}
	});
});