define(function () {
	app.registerController('CreateTaskCtrl', ['$scope', '$http', 'Template', 'Project', function ($scope, $http, Template, Project) {
		console.log(Template);
		$scope.task = {};

		$scope.run = function (task, dryRun) {
			task.template_id = Template.id;

			var params = angular.copy(task);
			if (dryRun) {
				params.dry_run = true;
			}
			$http.post(Project.getURL() + '/tasks', params).then(function (t) {
				$scope.$close(t.data);
			}).catch(function (response) {
				swal('Error', 'error launching task: HTTP ' + response.status, 'error');
			});
		}
	}]);

	app.registerController('TaskCtrl', ['$scope', '$http', function ($scope, $http) {
		$scope.raw = false;
		$scope.task = $scope.task;
		var logData = [];
		var onDestroy = [];

		onDestroy.push($scope.$on('task.log', function (evt, data) {
			var o = data.output + '\n';
			var d = moment(data.time);
			if (!$scope.raw) {
				o = d.format('HH:mm:ss') + ': ' + o;
			}

			if ($scope.task.id !== data.task_id) {
				return;
			}

			for (var i = 0; i < logData.length; i++) {
				if (d.isAfter(logData[i].time)) {
					// too far -- no point scanning rest of data as its in chronological order
					break;
				}

				if (d.isSame(logData[i].time) && data.output == logData[i].output) {
					return;
				}
			}

			$scope.output_formatted += o;
			if (!$scope.$$phase) $scope.$digest();
		}));

		onDestroy.push($scope.$on('task.update', function (evt, data) {
			$scope.task.status = data.status;
			$scope.task.start = data.start;
			$scope.task.end = data.end;

			if (!$scope.$$phase) $scope.$digest();
		}));

		$scope.reload = function () {
			$http.get($scope.project.getURL() + '/tasks/' + $scope.task.id + '/output')
			.then(function (output) {
				logData = output.data;
				var out = [];
				output.data.forEach(function (o) {
					var pre = '';
					if (!$scope.raw) pre = moment(o.time).format('HH:mm:ss') + ': ';

					out.push(pre + o.output);
				});

				$scope.output_formatted = out.join('\n') + '\n';
			});
			if ($scope.task.user_id) {
				$http.get('/users/' + $scope.task.user_id)
				.then(function (output) {
					$scope.task.user_name = output.data.name;
				});
			}
		}

		$scope.remove = function () {
			$http.delete($scope.project.getURL() + '/tasks/' + $scope.task.id)
			.then(function () {
				$scope.$close();
			}).catch(function () {
				swal("Error", 'Could not delete task', 'error');
			});
		}

		$scope.$watch('raw', function () {
			$scope.reload();
		});

		$scope.$on('$destroy', function () {
			logData = null;
			onDestroy.forEach(function (f) {
				f();
			});
		});
	}]);
});
