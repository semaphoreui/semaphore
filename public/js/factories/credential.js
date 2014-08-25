define(['app'], function (app) {
	app.registerFactory('Credential', ['$http', function ($http) {
		var Credential = function (id) {
			if (!id) {
				return;
			}
			
			this.id = id;
		}

		Credential.prototype.save = function () {
			return $http.put('/credential/'+this.data._id, this.data);
		}

		Credential.prototype.add = function () {
			return $http.post('/credentials', this.data);
		}

		Credential.prototype.delete = function () {
			return $http.delete('/credential/'+this.data._id);	
		}

		Credential.prototype.get = function () {
			return $http.get('/credential/'+this.id);
		}

		return Credential;
	}])
})