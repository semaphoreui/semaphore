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
	}]);
});