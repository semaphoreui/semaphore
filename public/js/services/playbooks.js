define([
	'app'
], function(app) {
	app.registerService('playbooks', function($http, $rootScope) {
		var self = this;

		self.getPlaybooks = function(cb) {
			$http.get('/playbooks').success(function(data) {
				$rootScope.playbooks = self.playbooks = data;
				
				cb();
			});
		}
	});
});