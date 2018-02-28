define([
	'app',
	'factories/hostgroup'
], function(app) {
	app.registerService('hostgroups', ['$http', 'HostGroup', function($http, HostGroup) {
		var self = this;

		self.get = function(playbook, cb) {
			$http.get('/playbook/'+playbook.data._id+'/hostgroups').then(function(response) {
				self.hostgroups = [];

				for (var i = 0; i < response.data.length; i++) {
					var g = new HostGroup();
					g.data = response.data[i];
					
					g.getHosts(playbook);
					self.hostgroups.push(g);
				}
				
				cb();
			});
		}
	}]);
});