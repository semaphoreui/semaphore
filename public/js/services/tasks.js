define([
	'app',
	'factories/task'
], function(app) {
	app.registerService('tasks', ['$http', 'Task', function($http, Task) {
		var self = this;

		self.get = function(playbook, cb) {
			$http.get('/playbook/'+playbook.data._id+'/tasks').success(function(data) {
				self.tasks = [];

				for (var i = 0; i < data.length; i++) {
					var task = new Task();
					task.data = data[i];
					
					self.tasks.push(task);
				}
				
				cb();
			});
		}
	}]);
});