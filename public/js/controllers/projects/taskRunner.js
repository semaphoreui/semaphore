define(function () {
	app.registerController('CreateTaskCtrl', ['$scope', '$http', 'Template', 'Project', function ($scope, $http, Template, Project) {
		console.log(Template);
		$scope.task = {};

		$scope.run = function (task) {
			task.template_id = Template.id;

			$http.post(Project.getURL() + '/tasks', task).success(function (t) {
				$scope.$close(t);
			}).error(function (_, status) {
				swal('Error', 'error launching task: HTTP ' + status, 'error');
			});
		}
	}]);

	app.registerController('TaskCtrl', ['$scope', '$http', function ($scope, $http) {
		$scope.$on('remote.log', function (evt, data) {
			console.log('data');
			$scope.output_formatted += moment(data.time).format('HH:mm:ss') + ': ' + data.output + '\n';

			if (!$scope.$$phase) $scope.$digest();
		});

		$http.get($scope.project.getURL() + '/tasks/' + $scope.task.id + '/output')
		.success(function (output) {
			var out = [];
			output.forEach(function (o) {
				out.push(moment(o.time).format('HH:mm:ss') + ': ' + o.output);
			});

			$scope.output_formatted = out.join('\n');
		});
	}]);
});