define(function () {
	app.registerController('UserCtrl', ['$scope', '$http', '$uibModal', '$rootScope', 'user', '$state', 'SweetAlert', function ($scope, $http, $modal, $rootScope, user, $state, SweetAlert) {
		$scope.user = user.data;
		$scope.is_self = $scope.user.id == $rootScope.user.id;

		$scope.updatePassword = function (pwd) {
			$http.post('/users/' + $scope.user.id + '/password', {
				password: pwd
			}).then(function () {
				SweetAlert.swal('OK', 'User profile & password were updated.');
			}).catch(function (response) {
				SweetAlert.swal('Error', 'Setting password failed, API responded with HTTP ' + response.status, 'error');
			});
		}

		$scope.updateUser = function () {
			var pwd = $scope.user.password;

			$http.put('/users/' + $scope.user.id, $scope.user).then(function () {
				if ($rootScope.user.id == $scope.user.id) {
					$rootScope.user = $scope.user;
				}

				if (pwd && pwd.length > 0) {
					$scope.updatePassword(pwd);
					return;
				}

				SweetAlert.swal('OK', 'User has been updated!');
			}).catch(function (response) {
				SweetAlert.swal('Error', 'User profile could not be updated: ' + response.status, 'error');
			});
		}

		$scope.deleteUser = function () {
			$http.delete('/users/' + $scope.user.id).then(function () {
				$state.go('users.list');
			}).catch(function (response) {
				SweetAlert.swal('Error', 'User could not be deleted! ' + response.status, 'error');
			});
		}
	}]);
});
