define([
	'app'
], function(app) {
	app.registerController('SignInCtrl', ['$scope', '$rootScope', '$http', '$state', function($scope, $rootScope, $http, $state) {
		$scope.status = "";
		$scope.user = {
			auth: "",
			password: ""
		};
	
		$scope.authenticate = function(user) {
			$scope.status = "Authenticating..";
		
			var pwd = user.password;
			user.password = "";

			$http.post('/auth/password', {
				auth: user.auth,
				password: pwd
			}).success(function(data, status) {
				$scope.status = "Login Successful";
				window.location = "/";
			}).error(function (data, status) {
				if (status == 400) {
					// Login Failed
					$scope.status = data.message;
					
					return;
				}

				$scope.status = status + ' Request Failed. Try again later.';
			});
		}
	}]);
});