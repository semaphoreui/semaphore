define(['app'], function (app) {
	app.registerFactory('Task', ['$http', function ($http) {
		var Task = function (id) {
			if (!id) {
				return;
			}
			
			this.id = id;
		}

		Task.prototype.delete = function (playbook, job) {
			return $http.delete('/playbook/'+playbook.data._id+'/job/'+job._id+'/task/'+this.data._id);
		}

		Task.prototype.get = function () {
			return $http.get('/task/'+this.id);
		}

		return Task;
	}])
})