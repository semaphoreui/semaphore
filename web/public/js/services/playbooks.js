define([
	'app'
], function(app) {
	app.registerService('playbooks', function($http, $rootScope) {
		var self = this;

		self.getPlaybooks = function(cb) {
			$http.get('/playbooks').then(function(response) {
				$rootScope.playbooks = self.playbooks = response.data;
				
				cb();
			});
		}
	});
});