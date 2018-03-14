define(['app'], function (app) {
	app.registerFactory('Playbook', ['$http', function ($http) {
		var Playbook = function (id, cb) {
			if (!id) {
				return;
			}
			
			this.id = id;

			this.get(cb);
		}

		Playbook.prototype.save = function () {
			return $http.put('/playbook/'+this.data._id, this.data);
		}

		Playbook.prototype.add = function () {
			return $http.post('/playbooks', this.data);
		}

		Playbook.prototype.delete = function () {
			return $http.delete('/playbook/'+this.data._id);	
		}

		Playbook.prototype.get = function (cb) {
			var self = this;

			$http.get('/playbook/'+this.id)
			.then(function (response) {
				self.data = response.data;
				cb();
			})
			.then(function (response) {
				cb(response.data, response.status);
			});
		}

		Playbook.prototype.getHostGroups = function (cb) {
			$http.get('/playbook/'+this.data._id+'/hosts')
			.then(function (response) {
				
				self.hosts = response.data;
				cb();
			})
			.error(function (response) {
				cb(response.data, response.status);
			});
		}

		return Playbook;
	}]);
});