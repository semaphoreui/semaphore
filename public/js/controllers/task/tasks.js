define([
	'app',
	'jquery',
	'socketio',
	'services/tasks',
	'factories/task'
], function(app, $, io) {
	var socket = io();

	app.registerController('TasksCtrl', ['$scope', 'tasks', 'Task', function($scope, tasks, Task) {
		$scope.tasks = tasks;
		
		tasks.get($scope.playbook, function () {
		});

		$scope.onPlaybookUpdate = function (data) {
			if (data.playbook_id != $scope.playbook.data._id) return;

			var found = false;

			for (var i = 0; i < $scope.tasks.tasks.length; i++) {
				var task = $scope.tasks.tasks[i];

				if (task.data._id == data.task_id) {
					task.data = data.task;
					found = true;

					break;
				}
			}

			if (!found) {
				// add task??
				$scope.tasks.tasks.splice(0, 0, new Task());
				$scope.tasks.tasks[0].data = data.task;
			}

			if (!$scope.$$phase) {
				$scope.$digest();
			}
		};

		socket.on('playbook.update', $scope.onPlaybookUpdate);
		$scope.$on('$destroy', function () {
			// prevents memory leaks..
			socket.removeListener('playbook.update', $scope.onPlaybookUpdate);
		});
	}]);
});