define([
	'app',
	'factories/task'
], function(app) {
	app.registerService('tasks', ['$http', 'Task', function($http, Task) {
		var self = this;

		self.get = function(playbook, cb) {
			$http.get('/playbook/'+playbook.data._id+'/tasks').then(function(response) {
				self.tasks = [];

				for (var i = 0; i < response.data.length; i++) {
					var task = new Task();
					task.data = response.data[i];
					
					self.tasks.push(task);
				}
				
				cb();
			});
		}
	}]);
});