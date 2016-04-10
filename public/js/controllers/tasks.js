define([
	'app',
	'jquery',
	'socketio',
	'services/tasks',
	'factories/task'
], function(app, $, io) {
	var socket = io();

	app.registerController('PlaybookTasksCtrl', ['$scope', 'tasks', 'Task', function($scope, tasks, Task) {
		$scope.tasks = tasks;

		tasks.get($scope.playbook, function () {
		});

		$scope.onPlaybookUpdate = function (data) {
			if (data.playbook_id != $scope.playbook.data._id) return;

			var task = $scope.findTask(data.task_id);

			if (!task) {
				// add task??
				$scope.tasks.tasks.splice(0, 0, new Task());
				$scope.tasks.tasks[0].data = data.task;
			} else {
				task.data = data.task;
			}

			if (!$scope.$$phase) {
				$scope.$digest();
			}
		};

		$scope.onSocketOutput = function (data) {
			if (data.playbook_id != $scope.playbook.data._id) return;

			var task = $scope.findTask(data.task_id);
			if (!task) return;

			if (!task.data.output) {
				task.data.output = "";
			}

			task.data.output += data.output;

			if (!$scope.$$phase) {
				$scope.$digest();
			}
		}

		$scope.findTask = function (task_id) {
			for (var i = 0; i < $scope.tasks.tasks.length; i++) {
				var task = $scope.tasks.tasks[i];

				if (task.data._id == task_id) {
					return task;
				}
			}

			return null;
		}

		$scope.removeTask = function (task) {
			task.delete($scope.playbook, task.data.job)
			.success(function () {
				$scope.status = 'Task Deleted';

				for (var i = 0; i < $scope.tasks.tasks.length; i++) {
					if ($scope.tasks.tasks[i] == task) {
						$scope.tasks.tasks.splice(i, 1);
						break;
					}
				}
			})
			.error(function (data, status) {
				$scope.status = 'Task Failed to Delete ('+status+'): '+data;
			});
		}

		$scope.openOutput = function (task) {
			$scope.openTask = task;
			$('#taskOutput').modal('show');
			$('#taskOutput').on('hide.bs.modal', $scope.closeOutput);
		}

		$scope.closeOutput = function () {
			$scope.openTask = null;

			$('#taskOutput').off('hide.bs.modal', $scope.closeOutput);
			$('#taskOutput').modal('hide');
		}

		socket.on('playbook.update', $scope.onPlaybookUpdate);
		socket.on('playbook.output', $scope.onSocketOutput);
		$scope.$on('$destroy', function () {
			// prevents memory leaks..
			socket.removeListener('playbook.update', $scope.onPlaybookUpdate);
			socket.removeListener('playbook.output', $scope.onSocketOutput);
		});
	}]);
});