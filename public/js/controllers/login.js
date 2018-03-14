define(function () {
	app.registerController('SignInCtrl', ['$scope', '$rootScope', '$http', '$state', function ($scope, $rootScope, $http, $state) {
		$scope.status = "";
		$scope.user = {
			auth: "",
			password: ""
		};

		$scope.authenticate = function (user) {
			$scope.status = "Authenticating..";

			var pwd = user.password;
			user.password = "";

			$http.post('/auth/login', {
				auth: user.auth,
				password: pwd
			}).then(function (response) {
				$scope.status = "Login Successful";
				window.location = document.baseURI;
			}).catch(function (response) {
				if (response.status === 400) {
					// Login Failed
					$scope.status = response.data.message;
					if (!response.data.message) {
						$scope.status = "Invalid login";
					}

					return;
				}

				$scope.status = response.status + ' Request Failed. Try again later.';
			});
		}
	}]);
});