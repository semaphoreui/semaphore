define(['app'], function (app) {
	app.registerFactory('Identity', ['$http', function ($http) {
		var Model = function (id) {
			if (!id) {
				return;
			}
			
			this.id = id;
		}

		Model.prototype.save = function () {
			return $http.put('/identity/'+this.data._id, this.data);
		}

		Model.prototype.add = function () {
			return $http.post('/identities', this.data);
		}

		Model.prototype.delete = function () {
			return $http.delete('/identity/'+this.data._id);	
		}

		Model.prototype.get = function () {
			return $http.get('/identity/'+this.id);
		}

		return Model;
	}])
})