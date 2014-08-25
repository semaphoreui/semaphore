define(['app'], function (app) {
	app.registerFactory('Playbook', ['$http', function ($http) {
		var Playbook = function (id) {
			if (!id) {
				return;
			}
			
			this.id = id;
		}

		Playbook.prototype.save = function () {
			return $http.put('/credential/'+this.data._id, this.data);
		}

		Playbook.prototype.add = function () {
			return $http.post('/credentials', this.data);
		}

		Playbook.prototype.delete = function () {
			return $http.delete('/credential/'+this.data._id);	
		}

		Playbook.prototype.get = function (cb) {
			return $http.get('/credential/'+this.id);
		}

		return Playbook;
	}])
})