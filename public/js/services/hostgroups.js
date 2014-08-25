define([
	'app',
	'factories/HostGroup'
], function(app) {
	app.registerService('hostgroups', ['$http', 'HostGroup', function($http, HostGroup) {
		var self = this;

		self.get = function(playbook, cb) {
			$http.get('/playbook/'+playbook.data._id+'/hostgroups').success(function(data) {
				self.hostgroups = [];

				for (var i = 0; i < data.length; i++) {
					var g = new HostGroup();
					g.data = data[i];
					
					g.getHosts(playbook);
					self.hostgroups.push(g);
				}
				
				cb();
			});
		}
	}]);
});