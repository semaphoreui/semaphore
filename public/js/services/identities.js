define([
	'app'
], function(app) {
	app.registerService('identities', function($http, $rootScope) {
		var self = this;

		self.getIdentities = function(cb) {
			$http.get('/identities').then(function(response) {
				self.identities = response.data;
				
				cb();
			});
		}
	});
});