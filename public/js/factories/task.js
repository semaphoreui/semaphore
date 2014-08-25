define(['app'], function (app) {
	app.registerFactory('Task', ['$http', function ($http) {
		var Task = function (id) {
			if (!id) {
				return;
			}
			
			this.id = id;
		}

		Task.prototype.save = function () {
			return $http.put('/task/'+this.data._id, this.data);
		}

		Task.prototype.add = function () {
			return $http.post('/tasks', this.data);
		}

		Task.prototype.delete = function () {
			return $http.delete('/task/'+this.data._id);	
		}

		Task.prototype.get = function () {
			return $http.get('/task/'+this.id);
		}

		return Task;
	}])
})