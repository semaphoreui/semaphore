define(['app'], function (app) {
	app.registerFactory('Host', ['$http', function ($http) {
		var Host = function (id) {
			if (!id) {
				return;
			}
			
			this.id = id;
		}

		Host.prototype.save = function (playbook, hostgroup) {
			return $http.put('/playbook/'+playbook.data._id+'/hostgroup/'+hostgroup.data._id+'/host/'+this.data._id, this.data);
		}

		Host.prototype.add = function (playbook, hostgroup) {
			return $http.post('/playbook/'+playbook.data._id+'/hostgroup/'+hostgroup.data._id+'/hosts', this.data);
		}

		Host.prototype.delete = function (playbook, hostgroup) {
			return $http.delete('/playbook/'+playbook.data._id+'/hostgroup/'+hostgroup.data._id+'/host/'+this.data._id);	
		}

		Host.prototype.get = function (playbook, hostgroup) {
			return $http.get('/playbook/'+playbook.data._id+'/hostgroup/'+hostgroup.data._id+'/host/'+this.id);
		}

		return Host;
	}])
})