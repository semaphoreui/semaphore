define(['app', 'factories/host'], function (app) {
	app.registerFactory('HostGroup', ['$http', 'Host', function ($http, Host) {
		var HostGroup = function (id) {
			if (!id) {
				return;
			}
			
			this.id = id;
		}

		HostGroup.prototype.save = function (playbook) {
			return $http.put('/playbook/'+playbook.data._id+'/hostgroup/'+this.data._id, this.data);
		}

		HostGroup.prototype.add = function (playbook) {
			return $http.post('/playbook/'+playbook.data._id+'/hostgroups', this.data);
		}

		HostGroup.prototype.delete = function (playbook) {
			return $http.delete('/playbook/'+playbook.data._id+'/hostgroup/'+this.data._id);	
		}

		HostGroup.prototype.get = function (playbook) {
			return $http.get('/playbook/'+playbook.data._id+'/hostgroup/'+this.id);
		}

		HostGroup.prototype.getHosts = function (playbook) {
			var self = this;

			$http.get('/playbook/'+playbook.data._id+'/hostgroup/'+this.data._id+'/hosts')
			.then(function (response) {
				self.hosts = [];

				for (var i = 0; i < response.data.length; i++) {
					var g = new Host();
					g.data = response.data[i];
					
					self.hosts.push(g);
				}
			})
		}

		return HostGroup;
	}])
})