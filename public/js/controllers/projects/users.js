define(function () {
	app.registerController('ProjectUsersCtrl', ['$scope', '$http', 'Project', function ($scope, $http, Project) {
		$scope.reload = function () {
			$http.get(Project.getURL() + '/users').success(function (users) {
				$scope.users = users;
			});
		}

		$scope.remove = function (user) {
			$http.delete(Project.getURL() + '/users/' + user.id).success(function () {
				$scope.reload();
			}).error(function () {
				swal('error', 'could not delete user..', 'error');
			});
		}

		$scope.reload();
	}]);
});