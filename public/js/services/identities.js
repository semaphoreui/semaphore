define([
	'app'
], function(app) {
	app.registerService('identities', function($http, $rootScope) {
		var self = this;

		self.getIdentities = function(cb) {
			$http.get('/identities').success(function(data) {
				self.identities = data;
				
				cb();
			});
		}
	});
});