define(['app'], function (app) {
	app.registerFactory('User', ['$http', function ($http) {
		var Model = function (id) {
			if (!id) {
				return;
			}

			this.id = id;
		}

		Model.prototype.save = function () {
			return $http.put('/user/'+this.data._id, this.data);
		}

		Model.prototype.add = function () {
			return $http.post('/users', this.data);
		}

		Model.prototype.delete = function () {
			return $http.delete('/user/'+this.data._id);
		}

		Model.prototype.get = function () {
			return $http.get('/user/'+this.id);
		}

		return Model;
	}])
})
