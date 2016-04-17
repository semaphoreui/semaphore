define(function () {
	app.registerController('CreateTaskCtrl', ['$scope', '$http', 'Template', 'Project', function ($scope, $http, Template, Project) {
		console.log(Template);
		$scope.task = {};

		$scope.run = function (task) {
			task.template_id = Template.id;

			$http.post(Project.getURL() + '/tasks', task).success(function (t) {
			}).error(function (_, status) {
				swal('Error', 'error launching task: HTTP ' + status, 'error');
			});
		}
	}]);

	app.registerController('TaskCtrl', ['$scope', '$http', function ($scope, $http) {
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